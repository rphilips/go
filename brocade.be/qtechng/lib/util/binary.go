package util

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

func RefreshBinary(alsoqtechng bool) (err error) {
	exe := qregistry.Registry["qtechng-exe"]
	if exe == "" {
		return
	}
	pexe, err := exec.LookPath(exe)
	if err != nil {
		return err
	}
	tmp, err := qfs.TempFile("", "qtechng-bin-")
	if err != nil {
		return err
	}

	if err != nil {
		return
	}
	if alsoqtechng {
		err = qfs.GetURL(qregistry.Registry["qtechng-url"], tmp, "scriptfile")
		if err != nil {
			return err
		}
		err = qfs.RefreshEXE(pexe, tmp)
		if err != nil {
			return err
		}
		if runtime.GOOS == "linux" && strings.ContainsAny(qregistry.Registry["qtechng-type"], "BP") {
			err = qfs.SetPathmode(pexe, "scriptfile")
			if err != nil {
				return err
			}
			perm := qfs.CalcPerm("rwxrwxr-x")
			err = os.Chmod(pexe, perm|os.ModeSetuid|os.ModeSetgid)
			if err != nil {
				return err
			}
		}
	}
	qfs.Rmpath(tmp)

	if runtime.GOOS == "windows" {

		tmp, err := qfs.TempFile("", "qtechng-bin-")
		if err != nil {
			return err
		}
		u := qregistry.Registry["qtechng-url"]
		k := strings.LastIndex(u, "/")
		if k < 0 {
			return nil
		}
		url := u[:k+1] + "qtechngw-windows-amd64"
		pexe = strings.ReplaceAll(pexe, "qtechng", "qtechngw")
		if !qfs.IsFile(pexe) {
			err = qfs.GetURL(url, tmp, "tempfile")
			if err != nil {
				return err
			}
			err = qfs.CopyFile(tmp, pexe, "script", false)
			if err != nil {
				return errors.New("copying `" + tmp + "` to `" + pexe + "`: " + err.Error())
			}
		}
		qfs.Rmpath(tmp)
	}

	return nil
}

func VSCode(dir string) (err error) {
	vsixes, err := qfs.Find(dir, []string{"*.vsix"}, false, true, false)
	if err != nil {
		return
	}
	if len(vsixes) == 0 {
		return
	}

	vscode := qregistry.Registry["vscode-exe"]
	if vscode == "" {
		return
	}
	pvscode, err := exec.LookPath(vscode)
	if err != nil {
		return
	}

	ext := path.Ext(pvscode)

	args := make([]string, 0)

	program := ""
	switch ext {
	case ".bat", ".BAT", ".cmd", ".CMD":
		args = append(args, "cmd.exe", "/C", pvscode)
		program, err = exec.LookPath("cmd.exe")
		if err != nil {
			return
		}
	default:
		args = append(args, pvscode)
		program = pvscode
	}
	for _, vsix := range vsixes {
		args = append(args, "--install-extension", vsix)
	}
	args = append(args, "--force")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   program,
		Args:   args,
		Dir:    dir,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	err = cmd.Run()

	// install tasks.json
	tasks := filepath.Join(dir, "tasks.json")
	if !qfs.Exists(tasks) {
		return
	}

	workdir := qregistry.Registry["qtechng-work-dir"]
	dotdir := filepath.Join(workdir, ".vscode")
	os.MkdirAll(dotdir, 0755)
	err = qfs.CopyFile(tasks, dotdir, "rw-rw-rw-", false)
	return
}
