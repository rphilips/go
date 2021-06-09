package cmd

import (
	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
)

var versionRebuildCmd = &cobra.Command{
	Use:   "rebuild version",
	Short: "rebuild the underlying infrastructure of a version",
	Long: `Rebuild recontsructs the underlying infrastructure of a version: the uniquenss of basenames and the objects
Note that there is no installation`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version rebuild 0.00",
	RunE:    versionRebuild,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	versionCmd.AddCommand(versionRebuildCmd)
}

func versionRebuild(cmd *cobra.Command, args []string) error {

	r := qserver.Canon(args[0])
	release, err := qserver.Release{}.New(r, false)
	if err != nil {
		err := &qerror.QError{
			Ref: []string{"rebuild.notexist"},
			Msg: []string{"version does not exist."},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent)
		return nil
	}

	release.ReInit()
	qpaths := release.QPaths()

	err = qsource.Rebuild("rebuildversion", release.String(), qpaths)
	msg := make(map[string]string)
	msg["status"] = "Rebuild FAILED"
	msg["previous"] = ""

	if err == nil {
		msg["status"] = "Rebuild SUCCESS"
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent)
	return nil
}
