package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var lockDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes a lock",
	Long: `First argument is the name of a lock. 
`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng lock delete mylock`,
	Run:     lockDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	lockCmd.AddCommand(lockDeleteCmd)
}

func lockDelete(cmd *cobra.Command, args []string) {
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
	rand.Seed(time.Now().UnixNano())
	rnd := strconv.FormatInt(rand.Int63n(100000000), 10)
	tempdir := locker + "." + rnd
	os.Rename(locker, tempdir)
	os.RemoveAll(tempdir)
	if qfs.IsDir(locker) {
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Cannot remove lock: %s", lock))
		return
	}
}
