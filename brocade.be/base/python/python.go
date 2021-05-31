package python

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"brocade.be/base/registry"
)

// Compile tests if a python script compiles:
//    If the script does not exist: returns false
//    If the script has syntax errors: returns false
//    If the script has no syntax errors: returns true
func Compile(scriptpy string, py3 bool) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	tdir := registry.Registry["scratch-dir"]
	pyexe := GetPython(py3)
	argums := []string{
		pyexe,
		"-c",
		"import sys; filename=sys.argv[1]; source = open(filename).read(); print('<start>'); compile(source, filename, 'exec'); print('<compile ok>')",
		scriptpy,
	}
	cmd := exec.Cmd{
		Path:   pyexe,
		Args:   argums,
		Dir:    tdir,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	cmd.Run()

	sout := stdout.String()
	if strings.Contains(sout, "<compile ok>") {
		return nil
	}
	return fmt.Errorf("compile `%s` with `%s`:\n%s", scriptpy, pyexe, stderr.String())
}

// Run a python script with arguments in a directory
func Run(scriptpy string, py3 bool, args []string, extra []string, cwd string) (sout string, serr string) {
	pyexe := GetPython(py3)

	argums := []string{
		pyexe,
		"-c",
	}

	script := strings.ReplaceAll(scriptpy, "\\", "\\\\")
	script = strings.ReplaceAll(script, "\"", "\\\"")
	extra = append(extra, "exec(open('"+script+"').read())")
	argums = append(argums, strings.Join(extra, "; "))
	argums = append(argums, args...)

	if cwd == "" {
		c, e := os.Getwd()
		if e != nil {
			serr = "Current Working Directory not found:\n\n" + e.Error()
			return
		}
		cwd = c
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   pyexe,
		Args:   argums,
		Env:    nil,
		Dir:    cwd,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Run()
	sout = stdout.String()
	serr = stderr.String()
	if err != nil {
		serr += "\n\nError in " + cwd + ":\n    " + pyexe + " " + strings.Join(argums[1:], " ") + "\n\n    " + err.Error()
	}
	return
}

// GetPython retrieves from the registry the right python executable
func GetPython(py3 bool) string {
	pyexe := registry.Registry["python-exe"]
	py3exe := registry.Registry["python3-exe"]
	py2exe := registry.Registry["python2-exe"]
	if py3 && py3exe != "" {
		pyexe = py3exe
	}
	if !py3 && py2exe != "" {
		pyexe = py2exe
	}
	pyexe, _ = exec.LookPath(pyexe)
	return pyexe
}
