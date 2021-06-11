package rst

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
func Check(scriptrst string, level string) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tdir := registry.Registry["scratch-dir"]
	rstexe, _ := exec.LookPath("rstcheck")
	if level == "" {
		level = "info"
	}
	argums := []string{
		rstexe,
		"--report",
		level,
		"--ignore-roles",
		"menucall",
		"--ignore-directives",
		"sample",
		scriptrst,
	}
	cmd := exec.Cmd{
		Path:   rstexe,
		Args:   argums,
		Dir:    tdir,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cmd.Run()

	sout := stdout.String()
	sout = strings.TrimSpace(sout)
	serr := stderr.String()
	serr = strings.TrimSpace(serr)
	if serr == "" && sout == "" {
		return nil
	}
	return fmt.Errorf("check `%s` with `%s`:\n\nstdout:%s\n\nstderr:%s", scriptrst, rstexe, sout, serr)
}
