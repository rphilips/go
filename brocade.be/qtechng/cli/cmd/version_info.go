package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
)

var versionInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Information about a version",
	Long:    `Command provides information about a version`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version info 5.10",
	RunE:    versionInfo,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"fill-version":      "yes",
		"complete":          "version",
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
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	ok, _ := release.Exists("")
	if !ok {
		err = fmt.Errorf("version `%s` does NOT exist", release.String())
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	msg := make(map[string]interface{})
	fs := release.FS()
	msg["basedir"] = release.Root()
	msg["sourcedir"], _ = fs.RealPath("/")
	msg["version"] = release.String()

	msg["objects"] = release.ObjectCount()
	msg["projects"] = release.ProjectCount()
	msg["sources"] = release.SourceCount()

	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
