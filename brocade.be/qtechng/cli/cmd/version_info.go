package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
)

var versionInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Information about a version",
	Long:    `Command provides information about a version`,
	Args:    cobra.MaximumNArgs(1),
	Example: "qtechng version info 5.10",
	RunE:    versionInfo,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"fill-version":      "yes",
	},
}

func init() {
	versionInfoCmd.Flags().BoolVar(&Fremote, "remote", false, "Execute on the remote server")
	versionCmd.AddCommand(versionInfoCmd)
}

func versionInfo(cmd *cobra.Command, args []string) error {
	r := ""
	if len(args) > 0 {
		r = args[0]
	} else {
		r = Fversion
	}

	release, err := qserver.Release{}.New(r, true)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}

	ok, _ := release.Exists("")
	if !ok {
		err = fmt.Errorf("version `%s` does NOT exist", release.String())
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}

	msg := make(map[string]string)
	fs := release.FS()
	msg["basedir"] = release.Root()
	msg["sources"], _ = fs.RealPath("/")
	msg["version"] = release.String()

	if strings.Contains(QtechType, "B") {
		if filepath.Base(msg["basedir"]) == "0.00" {
			msg["~status"] = "ACTIVE"
		} else {
			msg["~status"] = "CLOSED"
		}
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml)
	return nil
}
