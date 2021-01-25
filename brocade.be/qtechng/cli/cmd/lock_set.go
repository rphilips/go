package cmd

import (
	"fmt"
	"os"
	"path"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var lockSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets a lock",
	Long: `Argument is the name of a lock. 
If this lock is set, no command can run with this lock
`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng lock set mylock`,
	Run:     lockSet,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	lockCmd.AddCommand(lockSetCmd)
}

func lockSet(cmd *cobra.Command, args []string) {
	lock := args[0]
	lockdir := qregistry.Registry["lock-dir"]
	if lockdir == "" {
		lockdir = qregistry.Registry["scratch-dir"]
	}
	if lockdir == "" {
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Cannot find lock-dir in registry"))
		return
	}

	locker := path.Join(lockdir, "brocade_"+lock)
	err := os.Mkdir(locker, os.ModePerm)
	if err != nil {
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Cannot create lock: %s", err.Error()))
		return
	}
}
