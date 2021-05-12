// build +windows

// Gui wrapper for windows applications.
// All magic is in the linker
// Just rename wingo to the application of your choice

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {

	exe := filepath.Base(os.Args[0])
	exe = strings.TrimSuffix(exe, ".exe")
	exe = strings.TrimSuffix(exe, "w") + ".exe"
	pexe, _ := exec.LookPath(exe)
	exe = filepath.Base(pexe)
	if strings.TrimSuffix(exe, ".exe") == "wingo" {
		return
	}
	os.Args[0] = exe
	attr := &syscall.SysProcAttr{}
	attr.CreationFlags = 0x08000000
	attr.HideWindow = true

	cmd := exec.Cmd{
		Path:        pexe,
		Args:        os.Args,
		SysProcAttr: attr,
	}
	cmd.Run()
}
