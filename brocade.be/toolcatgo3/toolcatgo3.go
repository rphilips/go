// The program should be run as root
// it enables runnung other packages to run as root

package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"brocade.be/base/registry"
)

func main() {
	ruid := syscall.Getuid()
	if ruid == 0 {
		syscall.Setresgid(ruid, ruid, ruid)
		ruid = syscall.Getuid()
	}
	toolcat := filepath.Base(os.Args[0])
	toolcat = strings.TrimSuffix(toolcat, filepath.Ext(toolcat))

	work := GetPython(true)
	// if ruid != 0 {
	// 	work, _ = exec.LookPath("sudo")
	// }
	cmd := exec.Cmd{
		Path:   work,
		Args:   append(Compile(toolcat, ruid), os.Args[1:]...),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
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

// Compile finds additional atguments
func Compile(toolcat string, ruid int) []string {
	pyexe := GetPython(true)
	if ruid == 0 {
		pyexe = filepath.Base(pyexe)
	}
	argums := make([]string, 0)
	if ruid != 0 {
		argums = append(argums, "sudo")
	}
	argums = append(argums, pyexe, "-c", `import sys; from anet.toolcatng import toolcat; toolcat.launch("`+toolcat+`", sys.argv[1:])`)
	return argums
}
