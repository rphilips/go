package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
)

var versionNewCmd = &cobra.Command{
	Use:     "new",
	Short:   "Creates a new version",
	Long:    `Command creates a new version on the development server`,
	Args:    cobra.NoArgs,
	Example: "qtechng version new",
	RunE:    versionNew,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"always-remote":  "yes",
		"with-qtechtype": "B",
	},
}

func init() {
	versionCmd.AddCommand(versionNewCmd)
}

func versionNew(cmd *cobra.Command, args []string) error {

	r := "0.00"

	release, err := qserver.Release{}.New(r, false)
	if err != nil {
		Fmsg = qreport.Report(nil, nil, Fjq, Fyaml, Funquote)
		return nil
	}

	err = release.Init()
	if err != nil {
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml, Funquote)
		return nil
	}

	ok, _ := release.Exists()
	if ok {
		Fmsg = fmt.Sprintf("Version `%s` is created", release.String())
	} else {
		err = fmt.Errorf("version `%s` is NOT created", release.String())
	}
	Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml, Funquote)
	return nil
}
