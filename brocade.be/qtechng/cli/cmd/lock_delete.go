package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var lockDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a lock",
	Long: `First argument is the name of a lock that should be deleted.
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
		Fmsg = qreport.Report(nil, fmt.Errorf("cannot find lock-dir in registry"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return
	}

	locker := filepath.Join(lockdir, "brocade_"+lock)
	rand.Seed(time.Now().UnixNano())
	rnd := strconv.FormatInt(rand.Int63n(100000000), 10)
	tempdir := locker + "." + rnd
	os.Rename(locker, tempdir)
	os.RemoveAll(tempdir)
	if qfs.IsDir(locker) {
		Fmsg = qreport.Report(nil, fmt.Errorf("cannot remove lock: %s", lock), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return
	}
}
