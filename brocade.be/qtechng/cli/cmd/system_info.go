package cmd

import (
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
)

var systemInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "List system information",
	Long:    `List configuration of setup for use in qtechng`,
	Args:    cobra.NoArgs,
	Example: "  qtechng system info",
	RunE:    systemInfo,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
	},
}

func init() {
	systemCmd.AddCommand(systemInfoCmd)
}

func systemInfo(cmd *cobra.Command, args []string) error {

	msg := make(map[string]string)
	host, e := os.Hostname()
	if e == nil {
		msg["!!uname"] = host
	}
	msg["UID"] = FUID
	msg["GOMAXPROCS"] = strconv.Itoa(runtime.GOMAXPROCS(-1))

	regkeys := regMap()

	for key := range regkeys {
		msg[key] = qregistry.Registry[key]
	}

	Fmsg = qerror.ShowResult(msg, Fjq, nil)
	return nil
}
