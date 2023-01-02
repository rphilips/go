package log

import (
	"encoding/json"
	"path/filepath"
	"time"

	bfs "brocade.be/base/fs"
	btime "brocade.be/base/time"
	pregistry "brocade.be/pbladng/lib/registry"
)

func Logfile() string {
	warnprops := pregistry.Registry["warn"]
	x := warnprops.(map[string]any)
	warndir := x["dir"].(string)
	logfile := filepath.Join(warndir, "log.json")
	if !bfs.IsDir(warndir) {
		bfs.MkdirAll(warndir, "process")
		bfs.Store(logfile, "{}", "process")
	}
	return filepath.Join(warndir, "log.json")
}

func Fetch() (log map[string]string, err error) {
	logfile := Logfile()
	data, err := bfs.Fetch(logfile)
	if err != nil {
		return
	}
	log = make(map[string]string)
	err = json.Unmarshal(data, &log)
	return
}

func Store(log map[string]string) (err error) {
	logfile := Logfile()
	data, _ := json.MarshalIndent(log, "", "    ")
	return bfs.Store(logfile, data, "process")
}
func SetMark(tag string, value string) {
	log, err := Fetch()
	if err != nil {
		return
	}
	now := time.Now()
	log[tag] = value
	log[tag+"-timestamp"] = btime.StringTime(&now, "I")
	Store(log)
}

func GetMark(tag string) (value string, stamp string) {
	log, err := Fetch()
	if err != nil {
		return
	}
	value = log[tag]
	stamp = log[tag+"-timestamp"]
	return
}
