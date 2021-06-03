package php

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"brocade.be/base/registry"
)

// Compile tests if a phpthon script compiles:
//    If the script does not exist: returns false
//    If the script has syntax errors: returns false
//    If the script has no syntax errors: returns true
func Compile(scriptphp string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tdir := registry.Registry["scratch-dir"]
	phpexe, _ := exec.LookPath("php")
	argums := []string{
		phpexe,
		"-l",
		scriptphp,
	}
	cmd := exec.Cmd{
		Path:   phpexe,
		Args:   argums,
		Dir:    tdir,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cmd.Run()

	sout := stdout.String()
	if strings.Contains(sout, "No syntax errors detected in") {
		return nil
	}
	return fmt.Errorf("compile `%s` with `%s`:\n%s", scriptphp, phpexe, sout)
}
