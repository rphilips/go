package sync

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Sync synchronises files from one version to the other
// vsource is always a version on the devlopment server
// vtarget is on the current server.
// vtarget can only be empty on a non-development machine
// if vtarget is empty it reduces to the value of registry("brocade-release")
func Sync(vsource string, vtarget string, force bool) (changed []string, deleted []string, err error) {
	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") && strings.Contains(qtechType, "P") && vsource == vtarget {
		return
	}
	syncrun := qregistry.Registry["qtechng-sync-exe"]
	copyrun := qregistry.Registry["qtechng-copy-exe"]
	run := ""
	regvalue := ""
	if !strings.Contains(qtechType, "B") {
		if syncrun == "" {
			err = &qerror.QError{
				Ref: []string{"sync.version.sync.registry"},
				Msg: []string{"Registry value `qtechng-sync-exe` is missing"},
			}
			return
		}
		run = syncrun
		regvalue = "qtechng-sync-exe"
	}
	if strings.Contains(qtechType, "B") {
		if copyrun == "" {
			err = &qerror.QError{
				Ref: []string{"sync.version.copy.registry"},
				Msg: []string{"Registry value `qtechng-copy-exe` is missing"},
			}
			return
		}
		run = copyrun
		regvalue = "qtechng-copy-exe"
	}

	runparts := make([]string, 0)
	err = json.Unmarshal([]byte(run), &runparts)

	if err != nil {
		err = &qerror.QError{
			Ref: []string{"sync.registry.json"},
			Msg: []string{"Registry value `" + regvalue + "` is not JSON: `" + err.Error() + "`"},
		}
		return
	}

	current := qregistry.Registry["brocade-release"]
	if current == "" {
		err = &qerror.QError{
			Ref: []string{"sync.version.production"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		return
	}
	current = qserver.Canon(current)

	if strings.Contains(qtechType, "P") {
		if vsource == "" {
			vsource = current
		}
		if vtarget == "" {
			vtarget = vsource
		}
	}
	if !strings.Contains(qtechType, "P") && !strings.Contains(qtechType, "B") {
		if vsource != "0.00" {
			err = &qerror.QError{
				Ref: []string{"sync.version.development.source"},
				Msg: []string{"Version of source should be `0.00`"},
			}
			return
		}

		if vtarget == "" {
			err = &qerror.QError{
				Ref: []string{"sync.version.development.target"},
				Msg: []string{"Version of target should not be empty"},
			}
			return
		}
	}

	if vtarget == "0.00" || vtarget == "" {
		err = &qerror.QError{
			Ref: []string{"sync.version.production.target"},
			Msg: []string{"Version of target should never be `0.00` or empty"},
		}
		return
	}

	lowest := qutil.LowestVersion(current, vtarget)
	if current != lowest {
		err = &qerror.QError{
			Ref: []string{"sync.version.production.lowest"},
			Msg: []string{"The version of target should be higher "},
		}
	}

	if vsource != "0.00" && vtarget != vsource {
		err = &qerror.QError{
			Ref: []string{"sync.version.production.sourcetarget"},
			Msg: []string{"Version of source should be `0.00` or target should be equal to source"},
		}
		return
	}

	// real start of sync/copy

	runargs := make([]string, len(runparts))
	for i, arg := range runparts {
		arg = strings.ReplaceAll(arg, "{versiontarget}", vtarget)
		arg = strings.ReplaceAll(arg, "{versionsource}", vsource)
		runargs[i] = arg
	}

	inm := runargs[0]
	inm, _ = exec.LookPath(inm)

	var cmd *exec.Cmd
	if len(runargs) == 1 {
		cmd = exec.Command(inm)
	} else {
		cmd = exec.Command(inm, runargs[1:]...)
	}
	pipe, e := cmd.StdoutPipe()
	if e != nil {
		err = &qerror.QError{
			Ref: []string{"sync.exe.pipe.build"},
			Msg: []string{"Cannot open pipe to `" + regvalue + "`: `" + e.Error() + "`"},
		}
		return
	}
	cmd = qutil.Credential(cmd)
	err = cmd.Start()
	if err != nil {
		err = &qerror.QError{
			Ref: []string{"sync.exe.pipe.run"},
			Msg: []string{"Cannot run `" + regvalue + "`: `" + err.Error() + "`"},
		}
		return
	}
	reader := bufio.NewReader(pipe)

	changed = make([]string, 0)
	deleted = make([]string, 0)
	go func() {
		defer pipe.Close()
		line, err := reader.ReadString('\n')
		for err == nil {
			// >f+++++++++ source/date.txt
			switch {
			case strings.HasPrefix(line, ">f"):
				parts := strings.SplitN(line, " ", 2)
				if len(parts) != 2 {
					break
				}
				f := strings.TrimSpace(parts[1])
				f = filepath.ToSlash(f)
				if !strings.HasPrefix(f, "source/data") {
					break
				}
				f = strings.TrimPrefix(f, "source/data")
				if f == "" || f == "/" {
					break
				}
				changed = append(changed, f)
			case strings.HasPrefix(line, "*deleting"):
				line = strings.TrimPrefix(line, "*deleting")
				line = strings.TrimSpace(line)
				f := filepath.ToSlash(line)
				if !strings.HasPrefix(f, "source/data") {
					break
				}
				f = strings.TrimPrefix(f, "source/data")
				if f == "" || f == "/" {
					break
				}
				deleted = append(deleted, f)
			}
			line, err = reader.ReadString('\n')
		}
	}()

	cmd.Wait()

	return
}
