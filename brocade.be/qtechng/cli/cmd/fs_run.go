package cmd

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"

	qregistry "brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var fsRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs an executable",
	Long: `First argument is the name of a lock. 
If this lock is set, the program does not run.
If it is not set, the program does run and sets the lock
and cleans afterwards.
`,
	Args:    cobra.ExactArgs(2),
	Example: `qtechng fs run mylock docpublish -rebuild`,
	Run:     fsRun,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsRunCmd)
}

func fsRun(cmd *cobra.Command, args []string) {
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
		log.Fatalln("Cannot find lock-dir in registry")
	}

	locker := path.Join(lockdir, "brocade_"+lock)
	err := os.Mkdir(locker, os.ModePerm)
	if err != nil {
		log.Fatalln("Cannot create", locker)
	}
	unlock := func() {
		os.Rename(locker, locker+".bak")
		os.Remove(locker + ".bak")
	}

	rcmd := exec.Command(argums[0], argums[1:]...)
	go signalWatcher(ctx, rcmd, unlock)
	rcmd.Stdin = os.Stdin
	rcmd.Stdout = os.Stdout
	rcmd.Stderr = os.Stderr
	rcmd.Dir = Fcwd
	rcmd.SysProcAttr = &syscall.SysProcAttr{}
	rcmd.SysProcAttr.Setpgid = true
	err = rcmd.Run()
	unlock()
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		log.Fatalln("Unable to run command witch lock", args)
	}
}

func signalWatcher(ctx context.Context, cmd *exec.Cmd, unlock func()) {
	signalChan := make(chan os.Signal, 100)
	// Listen for all signals
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	unlock()
}
