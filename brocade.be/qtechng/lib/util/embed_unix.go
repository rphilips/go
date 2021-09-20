//go:build !windows
// +build !windows

package util

import (
	"os"
	"os/exec"
	"syscall"
)

func Credential(cmd *exec.Cmd) *exec.Cmd {
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Geteuid()), Gid: uint32(os.Getegid())}
	return cmd
}
