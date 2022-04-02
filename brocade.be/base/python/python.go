package python

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	qcredential "brocade.be/base/credential"
	"brocade.be/base/fs"
	"brocade.be/base/registry"
)

// Compile tests if a python script compiles:
//    If the script does not exist: returns false
//    If the script has syntax errors: returns false
//    If the script has no syntax errors: returns true
func Compile(scriptpy string, py3 bool, warnings bool, ignores []string) error {
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
	qcredential.Credential(&cmd)
	cmd.Run()

	sout := stdout.String()
	if !strings.Contains(sout, "<compile ok>") {
		return fmt.Errorf("compile `%s` with `%s`:\n%s", scriptpy, pyexe, stderr.String())
	}
	if !warnings || !py3 {
		return nil
	}

	if !strings.ContainsRune(registry.Registry["qtechng-type"], 'B') {
		return nil
	}

	cfg := filepath.Join(registry.Registry["scratch-dir"], "flake8.cfg")
	if !fs.IsFile(cfg) {
		qtechng, e := exec.LookPath(registry.Registry["qtechng-exe"])
		if qtechng == "" || e != nil {
			return nil
		}
		var stdout3 bytes.Buffer
		var stderr3 bytes.Buffer
		cmd := exec.Cmd{
			Path:   qtechng,
			Args:   []string{"qtechng", "source", "co", "/tools/scrutiny/flake8.cfg"},
			Dir:    registry.Registry["scratch-dir"],
			Stdout: &stdout3,
			Stderr: &stderr3,
		}
		e = cmd.Run()
		if e != nil {
			panic(e)
		}
		if !fs.IsFile(cfg) {
			return nil
		}
	}
	args := make([]string, 0)
	args = append(args, pyexe, "-m", "flake8")
	if len(ignores) != 0 {
		args = append(args, "--extend-ignore="+strings.Join(ignores, ","))
	}
	args = append(args, "--config="+cfg, scriptpy)
	var stdout2 bytes.Buffer
	var stderr2 bytes.Buffer
	cmd = exec.Cmd{
		Path:   pyexe,
		Args:   args,
		Dir:    tdir,
		Stdout: &stdout2,
		Stderr: &stderr2,
	}
	qcredential.Credential(&cmd)
	e := cmd.Run()
	sout = stdout2.String()
	if e != nil {
		sout = e.Error() + "\n" + sout
	}
	if strings.TrimSpace(sout) != "" {
		return fmt.Errorf("lint `%s` with `flake8`:\n%s", scriptpy, sout)
	}
	return nil
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

	// a, _ := json.MarshalIndent(argums, "", "    ")
	// fmt.Println(string(a))

	if cwd == "" {
		c, e := os.Getwd()
		if e != nil {
			serr = "Current Working Directory not found:\n\n" + e.Error()
			return
		}
		cwd = c
	}

	for _, env := range extra {
		key, value, found := strings.Cut(env, "=")
		if found && key != "" {
			value = strings.TrimLeft(value, "'")
			value = strings.TrimRight(value, "'")
			os.Setenv(key, value)
		}
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
	qcredential.Credential(&cmd)

	err := cmd.Run()

	for _, env := range extra {
		key, value, found := strings.Cut(env, "=")
		if found && key != "" {
			value = strings.TrimLeft(value, "'")
			value = strings.TrimRight(value, "'")
			os.Unsetenv(key)
		}
	}
	sout = stdout.String()
	serr = stderr.String()
	if err != nil {
		serr += "\n\nError in " + cwd + ":\n    " + pyexe + " " + strings.Join(argums[1:], " ") + "\n\n    " + err.Error()
	}
	return
}

// GetPython retrieves from the registry the right Python executable
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
