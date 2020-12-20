package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	frsync "github.com/zloylos/grsync"
)

var versionCopyCmd = &cobra.Command{
	Use:   "copy source target",
	Short: "Copies all files from one version to another",
	Long: `Works with two *existing* versions. The files from the first version are copied to the second version.
- The source directory of the second version should be empty
- Version control files are not copied`,
	Args:    cobra.ExactArgs(2),
	Example: "qtechng version copy 5.10 5.20",
	RunE:    versionCopy,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"always-remote":  "yes",
		"with-qtechtype": "BW",
	},
}

func init() {
	versionCmd.AddCommand(versionCopyCmd)
}

func versionCopy(cmd *cobra.Command, args []string) error {

	rsource := qserver.Canon(args[0])

	orsource, err := qserver.Release{}.New(rsource, true)

	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"version.copy.new.source"},
			Version: rsource,
			Msg:     []string{"Cannot instantiate version"},
		}
		Fmsg = qerror.ShowResult("", Fjq, e)
		return nil
	}

	rtarget := qserver.Canon(args[1])

	ortarget, err := qserver.Release{}.New(rtarget, true)
	fst := ortarget.FS()
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"version.copy.new.target"},
			Version: rtarget,
			Msg:     []string{"Cannot instantiate version"},
		}
		Fmsg = qerror.ShowResult("", Fjq, e)
		return nil
	}
	dirs := fst.Dir("/", false, true)
	if len(dirs) != 0 {
		e := &qerror.QError{
			Ref:     []string{"version.copy.empty.target"},
			Version: rtarget,
			Msg:     []string{"Target is not empty"},
		}
		Fmsg = qerror.ShowResult("", Fjq, e)
		return nil
	}

	source := orsource.Root()
	source += string(os.PathSeparator)

	destination := ortarget.Root()

	options := frsync.RsyncOptions{
		Checksum: true,
		Quiet:    true,
		//Verbose: true,
		Archive: true,
		Timeout: 600,
		Exclude: []string{
			path.Join("source", ".hg"),
			path.Join("source", ".git"),
		},
	}
	task := frsync.NewTask(
		source,
		destination,
		options,
	)

	if err := task.Run(); err != nil {
		e := &qerror.QError{
			Ref: []string{"version.copy.rsync"},
			Msg: []string{fmt.Sprintf("Could not rsync from`%s` to `%s`", source, destination)},
		}
		Fmsg = qerror.ShowResult("", Fjq, e)
		return nil
	}
	Fmsg = qerror.ShowResult(fmt.Sprintf("Copied `%s` to `%s`", source, destination), Fjq, nil)
	return nil
}
