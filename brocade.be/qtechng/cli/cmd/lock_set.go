package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var lockSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Sets a lock",
	Long: `First argument is the name of a lock.
Optional second argument is the number of seconds the lock remains valid.
Without second argument (or value of '0') the lock is valid for eternity 
If this lock is set, no command can run with this lock
`,
	Args: cobra.MinimumNArgs(1),
	Example: `qtechng lock set mylock 3600
	qtechng lock set mylock
	`,
	Run: lockSet,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	lockCmd.AddCommand(lockSetCmd)
}

func lockSet(cmd *cobra.Command, args []string) {
	lock := args[0]
	until := ""
	if len(args) > 1 {
		until = args[1]
	}

	locker := checkLock(lock, until)

	if locker == "" {
		Fmsg = qreport.Report(nil, fmt.Errorf("cannot create lock `%s`", lock), Fjq, Fyaml)
		return
	}
}

func checkLock(lock string, until string) (locker string) {
	lockdir := qregistry.Registry["lock-dir"]
	if lockdir == "" {
		lockdir = qregistry.Registry["scratch-dir"]
	}
	if lockdir == "" {
		return ""
	}
	ioffset, err := strconv.Atoi(until)
	if err == nil && ioffset > 0 {
		h := time.Now()
		h = h.Add(time.Second * time.Duration(ioffset))
		until = h.Format(time.RFC3339)
	}
	locker = filepath.Join(lockdir, "brocade_"+lock)
	untilfile := filepath.Join(locker, "until")
	h := time.Now()
	now := h.Format(time.RFC3339)
	b, err := os.ReadFile(untilfile)
	if err == nil {
		bs := string(b)
		if bs < now {
			os.RemoveAll(locker)
		}
	}
	err = os.Mkdir(locker, os.ModePerm)
	if err != nil {
		return ""
	}
	qfs.Store(untilfile, until, "process")
	return locker
}
