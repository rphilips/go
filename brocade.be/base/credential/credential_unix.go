//go:build !windows
// +build !windows

package credential

import (
	"os"
	"os/exec"
	"syscall"
)

func Credential(cmd *exec.Cmd) *exec.Cmd {
	if os.Geteuid() != os.Getuid() {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(os.Geteuid()), Gid: uint32(os.Getegid())}
	}
	return cmd
}
