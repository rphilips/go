package mumps

import (
	"bytes"
	"os"
	"os/exec"
)

const CREATE_NO_WINDOW = 0x08000000

func MUPIP(args []string, cwd string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	qexe := "mupip"
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	pexe, e := exec.LookPath(qexe)
	if e != nil {
		panic("MUPIP not found in PATH")
	}
	argums := []string{
		qexe,
	}
	argums = append(argums, args...)
	cmd := exec.Cmd{
		Path:   pexe,
		Args:   argums,
		Dir:    cwd,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Run()
	sout := stdout.String()
	serr := stderr.String()

	return sout, serr, err
}
