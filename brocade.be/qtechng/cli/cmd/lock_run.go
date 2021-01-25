package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"
	"time"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var lockRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs an executable",
	Long: `First argument is the name of a lock. 
If this lock is set, the program does not run.
If it is not set, the program does run and sets the lock
and cleans up afterwards.
`,
	Args:    cobra.ExactArgs(2),
	Example: `qtechng lock run mylock docpublish -rebuild`,
	Run:     lockRun,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	lockCmd.AddCommand(lockRunCmd)
}

func lockRun(cmd *cobra.Command, args []string) {
	lock := args[0]
	exe := args[1]

	argums := make([]string, 0)
	json.Unmarshal([]byte(exe), &argums)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lockdir := qregistry.Registry["lock-dir"]
	if lockdir == "" {
		lockdir = qregistry.Registry["scratch-dir"]
	}
	if lockdir == "" {
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Cannot find `lock-dir` in registry"))
		return
	}

	locker := path.Join(lockdir, "brocade_"+lock)
	err := os.Mkdir(locker, os.ModePerm)
	if err != nil {
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Cannot create lock: %s", err.Error()))
		return
	}
	unlock := func(cmd *exec.Cmd, code int) {
		rand.Seed(time.Now().UnixNano())
		rnd := strconv.FormatInt(rand.Int63n(100000000), 10)
		tempdir := locker + "." + rnd
		os.Rename(locker, tempdir)
		os.Remove(tempdir)
		if cmd != nil {
			cmd.Process.Kill()
			return
		}
		os.Exit(code)
	}

	rcmd := exec.Command(argums[0], argums[1:]...)
	go signalWatcher(ctx, rcmd, 1, unlock)
	rcmd.Stdin = os.Stdin
	rcmd.Stdout = os.Stdout
	rcmd.Stderr = os.Stderr
	rcmd.Dir = Fcwd
	rcmd.SysProcAttr = &syscall.SysProcAttr{}
	rcmd.SysProcAttr.Setpgid = true
	err = rcmd.Run()
	unlock(nil, 0)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		Fmsg = qerror.ShowResult(nil, Fjq, fmt.Errorf("Unable to run command"))
		return
	}
}

func signalWatcher(ctx context.Context, cmd *exec.Cmd, code int, unlock func(cmd *exec.Cmd, code int)) {
	signalChan := make(chan os.Signal, 100)
	// Listen for all signals
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	unlock(cmd, code)
}
