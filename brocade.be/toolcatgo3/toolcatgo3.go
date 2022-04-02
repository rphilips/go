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
	euid := syscall.Geteuid()

	rgid := syscall.Getgid()
	egid := syscall.Getegid()

	if ruid != euid {
		syscall.Setresuid(euid, euid, euid)
	}
	if rgid != egid {
		syscall.Setresgid(egid, egid, egid)
	}

	toolcat := filepath.Base(os.Args[0])
	toolcat = strings.TrimSuffix(toolcat, filepath.Ext(toolcat))

	pyexe := GetPython(true)

	cmd := exec.Cmd{
		Path:   pyexe,
		Args:   append(Compile(toolcat, pyexe), os.Args[1:]...),
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
		Stderr: os.Stderr,
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Geteuid()), Gid: uint32(os.Getegid())}

	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
	}
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

// Compile finds additional atguments
func Compile(toolcat string, pyexe string) []string {
	return []string{
		filepath.Base(pyexe),
		"-c",
		`import sys; from anet.toolcatng import toolcat; toolcat.launch("` + toolcat + `", sys.argv[1:])`,
	}
}
