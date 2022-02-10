package sync

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	qcredential "brocade.be/base/credential"
	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Sync synchronises files from one version to the other
// vsource is always a version on the devlopment server
// vtarget is on the current server.
// vtarget can only be empty on a non-development machine
// if vtarget is empty it reduces to the value of registry("brocade-release")
func Sync(vsource string, vtarget string, force bool, deepy bool) (changed []string, deleted []string, err error) {
	deep := true
	shallow := !deep
	//shallow := false

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
		shallow = false
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
	if vsource != vtarget {
		shallow = false
	}

	// real start of sync/copy

	syncmap := make(map[string][]string)
	files := make([]string, 0)
	if shallow {
		sout, _, err := qutil.QtechNG([]string{
			"version",
			"modified",
			vsource,
		}, []string{"$..DATA"}, false, "")
		if err != nil {
			shallow = false
			syncmap = nil
		}
		if !strings.HasPrefix(sout, "{") {
			syncmap = nil
			shallow = false
		} else {
			json.Unmarshal([]byte(sout), &syncmap)
			context := syncmap["context"]
			if len(context) == 0 {
				syncmap = nil
				shallow = false
			}
		}
		if shallow && len(syncmap) > 1 {
			for k, fils := range syncmap {
				if k == "context" {
					continue
				}
				files = append(files, fils...)
			}
		}

		if shallow {
			tmpfile, _ := qfs.TempFile("", "qsync-")
			qfs.Store(tmpfile, strings.Join(files, "\n"), "qtech")
			x := make([]string, 0)
			x = append(x, runparts[0], "--files-from="+tmpfile)
			runparts = append(x, runparts[1:]...)
		}

	}

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
	cmd = qcredential.Credential(cmd)
	out, e := cmd.Output()

	if e != nil {
		err = &qerror.QError{
			Ref: []string{"sync.exe.error"},
			Msg: []string{e.Error()},
		}
		return
	}
	sout := string(out)

	deleted = make([]string, 0)

	mchanged := make(map[string]bool)

	targetVersion, _ := qserver.Release{}.New(vtarget, true)
	for _, line := range strings.SplitN(sout, "\n", -1) {
		switch {
		case strings.HasPrefix(line, ">f"):
			parts := strings.SplitN(line, " ", 2)
			if len(parts) != 2 {
				break
			}
			f := strings.TrimSpace(parts[1])
			f = filepath.ToSlash(f)
			if strings.HasPrefix(f, "source/data") {
				f = strings.TrimPrefix(f, "source/data")
				if f == "" || f == "/" {
					break
				}
				mchanged[f] = true
				continue
			}
			if strings.HasPrefix(f, "object/") && strings.HasSuffix(f, "/obj.json") {
				fname := filepath.Join(qregistry.Registry["qtechng-repository-dir"], vtarget, f)
				blob, e := qfs.Fetch(fname)
				if e != nil {
					continue
				}
				var r map[string]interface{}
				e = json.Unmarshal(blob, &r)
				if e != nil {
					continue
				}
				id := r["id"].(string)
				if id == "" {
					continue
				}
				parts := strings.SplitN(f, "/", -1)
				if len(parts) < 2 {
					continue
				}
				ty := parts[1]
				if !strings.HasSuffix(ty, "4") {
					continue
				}
				deps, e := qobject.GetDependenciesDeep(targetVersion, ty+"_"+id)
				if e != nil {
					continue
				}
				for _, fils := range deps {
					for _, fil := range fils {
						if strings.HasPrefix(fil, "/") {
							mchanged[fil] = true
						}
					}
				}
			}

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
	}

	changed = make([]string, len(mchanged))
	count := 0
	for f := range mchanged {
		changed[count] = f
		count++
	}
	stamp := ""
	if shallow {
		context := syncmap["context"]
		if len(context) != 0 {
			stamp = context[0]
		}
	}

	if !shallow && regvalue == "qtechng-copy-exe" {
		now := time.Now()
		stamp = now.Format(time.RFC3339Nano)
	}
	if stamp != "" {
		m := make(map[string]string)
		m["timestamp"] = stamp
		b, _ := json.Marshal(m)
		release, _ := qserver.Release{}.New(vsource, false)
		tsf, _ := release.FS("/").RealPath("/admin/sync.json")
		qfs.Store(tsf, b, "qtech")
	}
	return
}
