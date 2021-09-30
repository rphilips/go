//go:build !windows
// +build !windows

package util

import (
	"os"
	"syscall"
)

func Setuid() {
	euid := os.Geteuid()
	if euid != os.Getuid() {
		syscall.Setuid(euid)
		egid := os.Getegid()
		syscall.Setgid(egid)
	}
}
