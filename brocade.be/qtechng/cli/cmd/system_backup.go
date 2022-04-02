package cmd

import (
	"bytes"
	"errors"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
)

var systemBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Pushes version control to backup",
	Long: `This command pushes the git data to a backup server

Setup the backup server:

Let us designate this server as *backup.anet.be*,
the development server is *dev.anet.be*

    - this server should be trusted by *dev.anet.be* for SSH
	  access.
	- on this server, work with an empty directory: */presto/qtechng*
	- in this directory (as user *root*)::

	      git init --bare

On *dev.anet.be*:

Let us work in a temporary directory: */library/tmp/git*

    - mkdir /library/tmp/git
	- cd /library/tmp/git
	- sudo git clone --bare /library/repository/0.00/source
	- sudo git push --mirror root@backup.anet.be:/presto/qtechng
	- cd ..
	- rm -rf /library/tmp/git
    - cd /library/repository/0.00/source
	- sudo git remote add backup root@backup.anet.be:/presto/qtechng
    - sudo git push --mirror backup

Check the backup on *dev.anet.be* with:

   - cd /presto/qtechng
   - git log`,
	Args:    cobra.NoArgs,
	Example: `qtechng system backup`,
	RunE:    systemBackup,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"with-qtechtype": "BW",
		"remote-allowed": "yes",
		"always-remote":  "yes",
	},
}

func init() {
	systemCmd.AddCommand(systemBackupCmd)
}

func systemBackup(cmd *cobra.Command, args []string) error {
	version, _ := qserver.Release{}.New("0.00", true)
	sourcedir, _ := version.FS("").RealPath("/source")
	git := "qtechng-git-exe"
	var err error
	if git != "" && !qfs.Exists(git) {
		git, err = exec.LookPath("git")
		if err != nil {
			git, err = exec.LookPath("git.exe")
		}
		if err != nil {
			git = ""
		}
	}
	if git == "" || !qfs.Exists(git) {
		Fmsg = qreport.Report("", errors.New("cannot find `git`)"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	argums := []string{path.Base(git), "push", "--mirror", "backup"}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	push := exec.Cmd{
		Path:   git,
		Args:   argums,
		Dir:    sourcedir,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	err = push.Run()
	sout := strings.TrimSpace(stdout.String())
	serr := strings.TrimSpace(stderr.String())

	if serr != "" {
		sout = sout + "\n\n" + serr
	}
	sout = strings.TrimSpace(sout)

	if err != nil {
		Fmsg = qreport.Report(sout, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(sout, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
