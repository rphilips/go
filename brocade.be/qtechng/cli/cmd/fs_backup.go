package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

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

	- If no arguments are given, the command asks for arguments.
	- The other arguments: at least one file or directory are to be specified.
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
	fsBackupCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsBackupCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsBackupCmd.Flags().StringVar(&Fbackupfile, "backup", "", "File with backup")
	fsBackupCmd.Flags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	fsCmd.AddCommand(fsBackupCmd)
}

func fsBackup(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	if Fbackupfile == "" {
		fmt.Print("Backupfile ? : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		Fbackupfile = text
	}
	if Fbackupfile == "" {
		return nil
	}
	Fbackupfile = qutil.AbsPath(Fbackupfile, Fcwd)
	if qfs.Exists(Fbackupfile) {
		Fmsg = qreport.Report("", fmt.Errorf("backupfile `%s` exists already", Fbackupfile), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	extra, recurse, patterns, utf8only, _ := qutil.AskArg(args, 0, !Frecurse, len(Fpattern) == 0, !Futf8only, false)

	if len(extra) != 0 {
		args = append(args, extra...)
		if recurse {
			Frecurse = true
		}
		if len(patterns) != 0 {
			Fpattern = patterns
		}
		if utf8only {
			Futf8only = true
		}
	}

	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-backup-invalid")
			return nil
		}
		msg := make(map[string][]string)
		msg["backup"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
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
