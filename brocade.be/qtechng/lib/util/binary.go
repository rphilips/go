package util

import (
	"bytes"
	"os/exec"
	"path"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

func RefreshBinary() (err error) {
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
	err = qfs.GetURL(qregistry.Registry["qtechng-url"], tmp, "tempfile")
	if err != nil {
		return err
	}

	err = qfs.RefreshEXE(pexe, tmp)
	if err != nil {
		return err
	}
	qfs.Rmpath(tmp)
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
	return

}
