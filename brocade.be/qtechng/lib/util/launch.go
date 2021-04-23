package util

import (
	"bytes"
	"os/exec"

	qregistry "brocade.be/base/registry"
)

func QtechNG(args []string, jsonpath string, yaml bool, cwd string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	qexe := qregistry.Registry["qtechng-exe"]
	pexe, _ := exec.LookPath(qexe)
	argums := []string{
		qexe,
	}
	argums = append(argums, args...)
	if jsonpath != "" {
		argums = append(argums, "--jsonpath="+jsonpath)
	}
	if yaml {
		argums = append(argums, "--yaml")
	}
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
