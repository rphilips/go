//go:build windows
// +build windows

package util

import "os/exec"

func Credential(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}
