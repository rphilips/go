package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
)

var versionNewCmd = &cobra.Command{
	Use:     "new",
	Short:   "Creates a new version",
	Long:    `Command creates a new version on the development server`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version new 5.10",
	RunE:    versionNew,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"always-remote":  "yes",
		"with-qtechtype": "BW",
	},
}

func init() {
	versionCmd.AddCommand(versionNewCmd)
}

func versionNew(cmd *cobra.Command, args []string) error {

	r := args[0]

	release, err := qserver.Release{}.New(r, false)
	if err != nil {
		Fmsg = qerror.ShowResult("", Fjq, nil, Fyaml)
		return nil
	}

	err = release.Init()
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err, Fyaml)
		return nil
	}

	ok, _ := release.Exists()
	if ok {
		Fmsg = fmt.Sprintf("Version `%s` is created", release.String())
	} else {
		err = fmt.Errorf("Version `%s` is NOT created", release.String())
	}
	Fmsg = qerror.ShowResult(Fmsg, Fjq, err, Fyaml)
	return nil
}
