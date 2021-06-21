package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var versionBackupCmd = &cobra.Command{
	Use:   "backup version",
	Short: "Backup of version",
	Long: `Backup is in tar format: it stores the content of *meta* en *source/data*
	The result is always brocade-{version}-{timestamp}.tar in the current directory.`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version backup 0.00",
	RunE:    versionBackup,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	versionCmd.AddCommand(versionBackupCmd)
}

func versionBackup(cmd *cobra.Command, args []string) error {
	h := time.Now()
	t := h.Format(time.RFC3339)[:19]
	t = strings.ReplaceAll(t, ":", "")
	t = strings.ReplaceAll(t, "-", "")
	r := qserver.Canon(args[0])
	release, _ := qserver.Release{}.New(r, true)
	ok, _ := release.Exists("/source/data")
	if !ok {
		err := &qerror.QError{
			Ref: []string{"backup.notexist"},
			Msg: []string{"version does not exist."},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	tarfile := qutil.AbsPath("brocade-"+r+"-"+t+".tar", Fcwd)
	err := release.Backup(tarfile)

	msg := make(map[string]string)
	msg["status"] = "Backup FAILED"
	if err == nil {
		msg["status"] = "Backup SUCCESS to `" + tarfile + "`"
		msg["backupfile"] = tarfile
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return err
}
