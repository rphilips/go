package cmd

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
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
	if strings.Contains(QtechType, "B") {
		msg["releases"] = qserver.Releases(5)
	}
	msg["UID"] = FUID
	msg["GOMAXPROCS"] = strconv.Itoa(runtime.GOMAXPROCS(-1))
	beol := []byte("\n")
	eol := ""
	for _, b := range beol {
		eol += strconv.Itoa(int(b))
	}
	msg["eol"] = eol

	regkeys := regMap()

	for key := range regkeys {
		msg[key] = qregistry.Registry[key]
	}

	if len(Fenv) != 0 {
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", -1)
			msg["env "+parts[0]] = os.Getenv(parts[0])
		}
	}

	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml)
	return nil
}
