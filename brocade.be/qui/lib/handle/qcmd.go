package handle

import (
	"os/exec"
)

func Qcmd(cmds []string) string {

	cmds = append(cmds, "--yaml")

	cmd := exec.Command("qtechng", cmds...)
	outerr, _ := cmd.CombinedOutput()

	return string(outerr)
}
