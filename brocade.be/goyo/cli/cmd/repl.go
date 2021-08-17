package cmd

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/spf13/cobra"
)

var replCmd = &cobra.Command{
	Use:     "repl",
	Short:   "REPL for M",
	Long:    `REPL for M`,
	Args:    cobra.ArbitraryArgs,
	Example: `goyo repl`,
	RunE:    repl,
}

func init() {
	rootCmd.AddCommand(replCmd)
}

func repl(cmd *cobra.Command, args []string) error {
	shell := ishell.New()
	if !Fsilent {
		greet(nil)

	}

	// handle "greet".
	shell.AddCmd(&ishell.Cmd{
		Name:    "greet",
		Aliases: []string{"hello"},
		Help:    "greet user",
		Func:    greet,
	})

	// when started with "exit" as first argument, assume non-interactive execution
	if len(args) > 0 && args[1] == "exit" {
		shell.Process(args[1:]...)
	} else {
		// start shell
		shell.Run()
		// teardown
		shell.Close()
	}

	return nil
}
