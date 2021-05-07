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

	qfs "brocade.be/base/fs"
	"golang.org/x/sys/windows/registry"
)

func main() {

	exe := filepath.Base(os.Args[0])
	exe = strings.TrimSuffix(exe, ".exe")
	exe = strings.TrimSuffix(exe, "w") + ".exe"
	pexe, _ := exec.LookPath(exe)
	exe = filepath.Base(pexe)
	if strings.TrimSuffix(exe, ".exe") == "wingo" {
		putty()
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

func putty() {
	// Computer\HKEY_CURRENT_USER\SOFTWARE\SimonTatham\PuTTY\Session
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\SimonTatham\PuTTY\Sessions`, registry.QUERY_VALUE)
	if err != nil {
		qfs.Store("C:\\Users\\rphilips\\msg2e.txt", err.Error(), "")
		return
	}
	defer k.Close()
	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		qfs.Store("C:\\Users\\rphilips\\msg3e.txt", err.Error(), "")
		return
	}
	qfs.Store("C:\\Users\\rphilips\\msg3.txt", strings.Join(subkeys, "\n"), "")

	subvals, err := k.ReadValueNames(-1)
	qfs.Store("C:\\Users\\rphilips\\msg4.txt", strings.Join(subvals, "\n"), "")
}
