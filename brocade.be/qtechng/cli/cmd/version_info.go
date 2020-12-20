package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
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
		Fmsg = qerror.ShowResult("", Fjq, err)
		return nil
	}

	ok, _ := release.Exists("")
	if !ok {
		err = fmt.Errorf("Version `%s` does NOT exist", release.String())
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	msg := make(map[string]string)
	fs := release.FS()
	msg["basedir"] = release.Root()
	msg["sources"], _ = fs.RealPath("/")
	msg["version"] = release.String()
	Fmsg = qerror.ShowResult(msg, Fjq, nil)
	return nil
}
