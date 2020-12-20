package cmd

import (
	"os"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
)

var systemCheckCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check system informatione",
	Long:    `Command allows for checking configuration/testing of setup for use in qtechng`,
	Args:    cobra.NoArgs,
	Example: "  qtechng system check",
	RunE:    systemCheck,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
	},
}

func init() {
	systemCmd.AddCommand(systemCheckCmd)
}

func systemCheck(cmd *cobra.Command, args []string) error {

	msg := make(map[string]string)
	host, e := os.Hostname()
	if e == nil {
		msg["!!uname"] = host
	}

	regkeys := regMap()
	for key, value := range regkeys {
		rvalue := qregistry.Registry[key]
		regex := value[1]
		if regex == "" {
			continue
		}
		re := regexp.MustCompile("^" + regex + "$")
		if !re.MatchString(rvalue) {
			msg[key] = "ERROR Invalid value `" + rvalue + "`"
			continue
		}
		check := value[2]
		if check == "" {
			msg[key] = "OK"
			continue
		}
		if check == "dir" && !qfs.IsDir(rvalue) {
			msg[key] = "ERROR Invalid directory `" + rvalue + "`"
			continue
		}
		if check == "exe" {
			where, e := exec.LookPath(rvalue)
			if where == "" || e != nil {
				msg[key] = "ERROR Cannot find `" + rvalue + "`"
				continue
			}
		}
		msg[key] = "OK"
	}
	Fmsg = qerror.ShowResult(msg, Fjq, nil)

	return nil
}
