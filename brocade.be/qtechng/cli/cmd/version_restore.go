package cmd

import (
	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
)

var versionRestoreCmd = &cobra.Command{
	Use:   "restore backup",
	Short: "Restore a version from backup",
	Long: `This command restores the contents of *meta* en *source/data* of a specific version.
The last argument should be a tar file built with the *qtechng version backup* command`,
	Args:    cobra.ExactArgs(2),
	Example: "qtechng version 0.00 backup.tar",
	RunE:    versionRestore,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

var Finit bool

func init() {
	versionCmd.AddCommand(versionRestoreCmd)
	versionRestoreCmd.Flags().BoolVar(&Finit, "init", false, "Initialises source/data and meta in version")
}

func versionRestore(cmd *cobra.Command, args []string) error {

	r := qserver.Canon(args[0])
	release, err := qserver.Release{}.New(r, true)
	if err != nil {
		err := &qerror.QError{
			Ref: []string{"restore.notexist"},
			Msg: []string{"version does not exist."},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	previous, err := release.Restore(args[1], Finit)
	msg := make(map[string]string)
	msg["status"] = "Backup Restore FAILED"
	msg["previous"] = ""

	if err == nil {
		msg["status"] = "Backup Restore SUCCESS"
	}
	if previous != "" {
		msg["status"] += " (backup of previous situation: `" + previous + "`)"
		msg["previous"] = previous
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
