package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	log "github.com/sirupsen/logrus"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

var Flogdir string

func init() {
	if Flogdir == "-" {
		return
	}
	if Flogdir == "" {
		Flogdir = qregistry.Registry["qtechng-log-dir"]
		if Flogdir == "-" {
			return
		}
	}
	if Flogdir == "" && qregistry.Registry["qtechng-type"] == "W" {
		Flogdir = "-"
		return
	}
	if Flogdir == "" && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'B') {
		Flogdir = filepath.Join(qregistry.Registry["qtechng-repository-dir"], "0.00", "log")
	}
	if Flogdir == "" && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'P') {
		release := qregistry.Registry["brocade-release"]
		if release == "" {
			Flogdir = "-"
			return
		}
		Flogdir = filepath.Join(qregistry.Registry["qtechng-repository-dir"], release, "log")
	}
	if Flogdir == "" {
		Flogdir = "-"
		return
	}

	h := time.Now()
	ux := strconv.FormatInt(h.UnixNano(), 16)
	if len(ux) < 16 {
		ux = "00000000000000000"[:16-len(ux)] + ux
	}
	dir := filepath.Join(Flogdir, h.Format(time.RFC3339)[:10])
	fname := filepath.Join(dir, ux+".log")
	logfile, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		Pack()
		qfs.MkdirAll(dir, "qtech")
		logfile, err = os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return
		}
	}
	log.SetOutput(logfile)
	log.SetFormatter(&log.JSONFormatter{})
}

func Log(buildtime string, uid string, version string, args []string) {
	if Flogdir == "-" {
		return
	}
	verb := ""
	if len(args) != 0 {
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				verb = arg
				break
			}
		}
	}
	log.WithFields(
		log.Fields{
			"buildtime": buildtime,
			"verb":      verb,
			"uid":       uid,
			"args":      args,
			"version":   version,
		},
	).Info("run")
}

func Recover(buildtime string, uid string, version string, args []string) {
	if Flogdir == "-" {
		return
	}
	r := recover()
	if r == nil {
		return
	}
	verb := ""
	if len(args) != 0 {
		for _, arg := range args {
			if !strings.HasPrefix(arg, "-") {
				verb = arg
				break
			}
		}
	}
	pan := fmt.Sprintf("Panic: %v,\n%s", r, debug.Stack())
	log.WithFields(
		log.Fields{
			"buildtime": buildtime,
			"verb":      verb,
			"uid":       uid,
			"args":      args,
			"version":   version,
			"panic":     pan,
		},
	).Error("run")

	os.Stderr.Write([]byte(pan))
	os.Exit(1)
}

func Pack() (string, error) {
	if Flogdir == "-" {
		return "registry `qtechng-log-dir` not set", nil
	}

	dirs, err := qfs.Find(Flogdir, []string{"[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]"}, true, false, true)

	if err != nil {
		return "", err
	}
	if len(dirs) == 0 {
		return "Logdir is `" + Flogdir + "`", nil
	}

	sort.Strings(dirs)

	for _, dir := range dirs {
		err := handleLogdir(dir)
		if err != nil {
			return "", err
		}
	}

	return "Logdir is `" + Flogdir + "`", nil
}

func handleLogdir(dir string) error {
	h := time.Now()
	if filepath.Base(dir) >= h.Format(time.RFC3339)[:10] {
		return nil
	}
	target := dir + ".log"
	logs := make([]interface{}, 0)
	if qfs.IsFile(target) {
		data, err := qfs.Fetch(target)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, &logs)
		if err != nil {
			return err
		}
	}
	files, err := qfs.Find(dir, []string{"*.log"}, false, true, false)
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		data, err := qfs.Fetch(file)
		if err != nil {
			return err
		}
		var x interface{}
		lines := strings.SplitN(string(data), "\n", -1)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			err = json.Unmarshal([]byte(line), &x)
			if err != nil {
				return err
			}
			logs = append(logs, x)
		}
	}
	if len(logs) != 0 {
		data, err := json.Marshal(logs)
		if err != nil {
			return err
		}
		err = qfs.Store(target, data, "qtech")
		if err != nil {
			return err
		}
	}
	qfs.Rmpath(dir)
	return nil
}

func Panic(when string) (infos []string, err error) {

	if Flogdir == "-" {
		return nil, nil
	}
	infos = make([]string, 0)
	files := make([]string, 0)
	panfile := filepath.Join(Flogdir, when+".log")
	if qfs.Exists(panfile) {
		files = append(files, panfile)
	} else {
		dir := filepath.Join(Flogdir, when)
		var err error
		files, err = qfs.Find(dir, []string{"*.log"}, false, true, false)
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	for _, file := range files {
		x, e := handleFilePanic(file)
		if e != nil {
			return nil, e
		}
		if len(x) == 0 {
			continue
		}
		infos = append(infos, x...)
	}
	if len(infos) == 0 {
		return nil, nil
	}
	fmt.Println("infos:", infos)
	return infos, nil

}

func handleFilePanic(logfile string) (info []string, err error) {
	data, err := qfs.Fetch(logfile)
	if err != nil {
		return
	}
	if len(data) == 0 {
		return
	}
	logs := make([]interface{}, 0)
	if data[0] == '[' {
		err = json.Unmarshal(data, &logs)
		if err != nil {
			return nil, err
		}
	} else {
		lines := strings.SplitN(string(data), "\n", -1)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var x interface{}
			err = json.Unmarshal([]byte(line), &x)
			if err != nil {
				return
			}
			logs = append(logs, x)
		}
	}
	for _, one := range logs {
		mone := one.(map[string]interface{})
		_, ok := mone["panic"]
		if !ok {
			continue
		}
		pan := mone["panic"].(string)
		pan = strings.TrimSpace(pan)
		if pan == "" {
			continue
		}
		delete(mone, "panic")
		show, err := json.MarshalIndent(mone, "", "    ")
		if err != nil {
			return nil, err
		}
		info = append(info, string(show)+"\n\n"+pan)
	}

	return
}
