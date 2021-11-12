package cmd

import (
	"fmt"
	"io"
	"log"
	"strings"

	qaction "brocade.be/goyo/action"
	qcompleter "brocade.be/goyo/lib/completer"
	qhistory "brocade.be/goyo/lib/history"
	qutil "brocade.be/goyo/lib/util"
	qliner "github.com/peterh/liner"
	"github.com/spf13/cobra"
)

var Defaults map[string]string

var replCmd = &cobra.Command{
	Use:     "repl",
	Short:   "REPL for YottaDB",
	Long:    `REPL forYottaDB`,
	Args:    cobra.NoArgs,
	Example: `goyo repl`,
	RunE:    repl,
}

func init() {
	Defaults = qutil.LoadDefaults()
	rootCmd.AddCommand(replCmd)
}

func finish(line *qliner.State) {
	qhistory.SaveHistory(line)
	line.Close()
	fmt.Println("Bye!")
}

var Fgloref string
var Fexec string
var Fvalue string
var Fline *qliner.State

func repl(cmd *cobra.Command, args []string) error {
	Fline := qliner.NewLiner()
	qhistory.LoadHistory(Fline)
	qcompleter.SetCompleter(Fline)
	Fline.SetCtrlCAborts(true)
	defer finish(Fline)
	fmt.Println("Please use `exit` to exit this program\nand `help` to get information on actions")

	for {
		action, err := Fline.Prompt("> ")

		if err == qliner.ErrPromptAborted {
			log.Print("Aborted")
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print("Error reading line: ", err)
			continue
		}
		if action == "" {
			continue
		}

		key, text := qutil.KeyText(action)
		key = strings.ToLower(key)
		if key == "bye" || key == "exit" || key == "quit" || key == "h" || key == "halt" {
			break
		}
		if key == "#" {
			Fline.AppendHistory(action)
			continue
		}
		if _, ok := qaction.Actions[key]; !ok {
			text = action
			key = "exec"
		}
		history := qaction.RunAction(key, text)
		for _, h := range history {
			if h != "" {
				Fline.AppendHistory(h)
			}
		}
		fmt.Println()
	}
	return nil
}
