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
	Short: "Back up a version",
	Long: `This command backs up a version. The backup is in tar format.
It stores the content of *meta* and *source/data*, and can be used to restore a version with *qtechng version restore*.
The result is always brocade-{version}-{timestamp}.tar in the current directory.`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version backup 0.00",
	RunE:    versionBackup,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

var Ftar = false
var Fsqlite = false

func init() {
	versionCmd.AddCommand(versionBackupCmd)
	versionBackupCmd.PersistentFlags().BoolVar(&Ftar, "tar", false, "Backup in tar format")
	versionBackupCmd.PersistentFlags().BoolVar(&Fsqlite, "sqlite", false, "Backup in sqlite format")
}

func versionBackup(cmd *cobra.Command, args []string) error {
	if !Ftar && !Fsqlite {
		Ftar = true
		Fsqlite = true
	}
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
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	tarfile := ""
	sqlitefile := ""
	var err error = nil
	if Ftar {
		tarfile = qutil.AbsPath("brocade-"+r+"-"+t+".tar", Fcwd)
		err = release.TarBackup(tarfile)
	}
	if err == nil && Fsqlite {
		sqlitefile = qutil.AbsPath("brocade-"+r+"-"+t+".sqlite", Fcwd)
		err = release.SqliteBackup(sqlitefile)
	}

	msg := make(map[string]string)
	msg["status"] = "Backup FAILED"
	if err == nil {
		msg["status"] = "Backup SUCCESS to `" + tarfile + "`"
		if tarfile != "" {
			msg["backupfile"] = tarfile
		}
		if sqlitefile != "" {
			msg["sqlitefile"] = sqlitefile
		}
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return err
}
