package rst

import (
	"fmt"
	"os/exec"

	qpy "brocade.be/base/python"

	qregistry "brocade.be/base/registry"
)

// Compile tests if a phpthon script compiles:
//    If the script does not exist: returns false
//    If the script has syntax errors: returns false
//    If the script has no syntax errors: returns true
func Check(scriptrst string, level string) error {
	tdir := qregistry.Registry["scratch-dir"]
	rstexe, _ := exec.LookPath("rstcheck")

	if level == "" {
		level = "error"
	}
	argums := []string{
		scriptrst,
	}

	sout, serr := qpy.Run(rstexe, true, argums, nil, tdir)

	// cmd := exec.Cmd{
	// 	Path: rstexe,
	// 	Args: argums,
	// 	Dir:  tdir,
	// }
	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// stderr, err := cmd.StderrPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = cmd.Run()

	// extra := ""
	// if err != nil {
	// 	extra = err.Error()
	// }

	// out, _ := io.ReadAll(stdout)
	// sout := strings.TrimSpace(string(out))
	// er, _ := io.ReadAll(stderr)
	// serr := strings.TrimSpace(string(er))
	// if extra != "" {
	// 	serr = extra + "\n" + serr
	// }
	if serr == "" && sout == "" {
		return nil
	}
	return fmt.Errorf("check `%s` with `%s`:\n\nstdout:%s\n\nstderr:%s", scriptrst, rstexe, sout, serr)
}
