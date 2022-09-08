package tools

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Launch[T any | string](params []T, keys map[string]string, cwd string, stderr bool) (output []byte, err error) {
	if len(params) == 0 {
		err = fmt.Errorf("missing name of executable")
		return
	}
	parms := make([]string, len(params))
	for i, param := range params {
		parms[i] = fmt.Sprintf("%v", param)
	}

	exe, err := exec.LookPath(parms[0])
	if err != nil {
		return
	}

	args := make([]string, 0)
	for _, arg := range parms[1:] {
		for key, value := range keys {
			arg = strings.ReplaceAll(arg, "{"+key+"}", value)
		}
		args = append(args, arg)
	}
	cmd := exec.Command(exe, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	if !stderr {
		cmd.Stderr, _ = os.Open(os.DevNull)
	}
	output, err = cmd.Output()
	return
}
