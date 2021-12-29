package cmd

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	qfs "brocade.be/base/fs"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup files",
	Long: `Backup files to a sqlite archive

With 'sqlite3' installed, you can verify the backup with:

    sqlite3 mybackup.sqlite -At

You can extract the backup with:

sqlite3 mybackup.sqlite -Ax


Some remarks:

    - With the '--ask' flag, you can interactively specify the arguments and flags
	- At least one file or directory are to be specified.
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content
	- The '--backup' flag contains the name of the backup file (relative to the current working directory)`,

	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs backup . --cwd=../catalografie --backup=backup.sqlite`,
	RunE:    fsBackup,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Fbackupfile = ""

func init() {
	fsBackupCmd.Flags().StringVar(&Fbackupfile, "backup", "", "File with backup")
	fsCmd.AddCommand(fsBackupCmd)
}

func fsBackup(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"backup::" + Fbackupfile,
			"files:backup",
			"recurse:backup,files:" + qutil.UnYes(Frecurse),
			"patterns:backup,files:",
			"utf8only:backup,files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fbackupfile = argums["backup"].(string)
	}
	if Fbackupfile == "" {
		Fmsg = qreport.Report(nil, errors.New("missing backup file"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-backupfile")
		return nil
	}
	Fbackupfile = qutil.AbsPath(Fbackupfile, Fcwd)
	if qfs.Exists(Fbackupfile) {
		Fmsg = qreport.Report("", fmt.Errorf("backupfile `%s` exists already", Fbackupfile), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-exists")
		return nil
	}

	if len(args) == 0 {
		Fmsg = qreport.Report("", fmt.Errorf("no files to backup"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-nofiles1")
		return nil
	}

	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)

	if err != nil {
		Ferrid = "fs-backup-glob"
		return err
	}

	if len(files) == 0 {
		Fmsg = qreport.Report("", fmt.Errorf("no files to backup"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-nofiles2")
		return nil
	}

	err = sqlitebackup(Fbackupfile, files)

	msg := make(map[string][]string)
	msg["backuped"] = files
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}

func sqlitebackup(backupfile string, files []string) (err error) {
	db, err := sql.Open("sqlite", backupfile)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err = db.Exec(`
		CREATE TABLE sqlar (
			name TEXT PRIMARY KEY,
			mode INT,
  			mtime INT,
  			sz INT,
  			data BLOB
		);`); err != nil {
		return err
	}

	stmt, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, sz, data) Values($1,$2,$3,$4,$5)")

	if err != nil {
		return fmt.Errorf("cannot prepare: %v", err)
	}

	defer stmt.Close()
	for _, name := range files {
		data, err := qfs.Fetch(name)
		if err != nil {
			return fmt.Errorf("cannot get content of `%s`: %v", name, err)
		}

		mt, err := qfs.GetMTime(name)
		if err != nil {
			return fmt.Errorf("cannot get mtime of `%s`: %v", name, err)
		}
		sz, err := qfs.GetSize(name)
		if err != nil {
			return fmt.Errorf("cannot get size of `%s`: %v", name, err)
		}
		mode, err := qfs.GetPerm(name)
		if err != nil {
			return fmt.Errorf("cannot get access permissions of `%s`: %v", name, err)
		}
		mtime := mt.Unix()
		_, err = stmt.Exec(name, uint32(mode), mtime, sz, data)
		if err != nil {
			return fmt.Errorf("cannot exec: %v", err)
		}
	}

	return nil

}
