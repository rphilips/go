//go:build windows
// +build windows

package credential

import "os/exec"

func Credential(cmd *exec.Cmd) *exec.Cmd {
	return cmd
}
