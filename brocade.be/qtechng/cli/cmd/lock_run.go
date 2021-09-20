package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var lockRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a program with a lock",
	Long: `The first argument is the name of a lock for a program.
The second argument is an estimation in seconds of the duration of this process.
If this lock is set, the program does not run.
If it is not set, the lock is set and the program runs.
Afterwards, the lock is deleted.
The third argument is the executable to run.
`,
	Args:    cobra.MinimumNArgs(3),
	Example: `qtechng lock run mylock 10 docpublish -rebuild`,
	Run:     lockRun,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	lockCmd.AddCommand(lockRunCmd)
}

func lockRun(cmd *cobra.Command, args []string) {
}

func LockRunner(args []string) {
	lock := args[0]
	until := args[1]
	exe := args[2]

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	locker := checkLock(lock, until)
	if locker == "" {
		Fmsg = qreport.Report(nil, fmt.Errorf("cannot obtain lock `%s`", lock), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return
	}

	unlock := func(cmd *exec.Cmd, code int) {
		rand.Seed(time.Now().UnixNano())
		rnd := strconv.FormatInt(rand.Int63n(100000000), 10)
		tempdir := locker + "." + rnd
		os.Rename(locker, tempdir)
		os.RemoveAll(tempdir)
		if cmd != nil {
			cmd.Process.Kill()
			return
		}
		os.Exit(code)
	}

	rcmd := exec.Command(exe, args[3:]...)
	go signalWatcher(ctx, rcmd, 1, unlock)
	rcmd.Stdin = os.Stdin
	rcmd.Stdout = os.Stdout
	rcmd.Stderr = os.Stderr
	rcmd.Dir = Fcwd
	rcmd = qutil.Credential(rcmd)
	err := rcmd.Run()
	unlock(nil, 0)
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		Fmsg = qreport.Report(nil, fmt.Errorf("unable to run command succesfully"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
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
