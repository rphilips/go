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
	fmt.Println("Please use `exit` to exit this program.")

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
		if key == "bye" || key == "exit" || key == "quit" {
			break
		}
		if key == "#" {
			Fline.AppendHistory(action)
			continue
		}
		if !qcompleter.IsAction(key) {
			Fline.AppendHistory(action)
			fmt.Println("?")
			continue
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

// }
// func completer(d prompt.Document) []prompt.Suggest {
// 	s := []prompt.Suggest{

// 		{Text: "users", Description: "Store the username and age"},
// 		{Text: "articles", Description: "Store the article text posted by user"},
// 		{Text: "comments", Description: "Store the text commented to articles"},
// 	}
// 	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
// }

// 	func main() {
// 		fmt.Println("Please select table.")
// 		t := prompt.Input("> ", completer)
// 		fmt.Println("You selected " + t)
// 	}
// 	// handle "greet".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "greet",
// 		Help: "greet user",
// 		Func: greet,
// 	})

// 	// handle "set".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "set",
// 		Help: "set var or global",
// 		Func: set,
// 	})

// 	// handle "cd".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "cd",
// 		Help: "cd directory (~ = home)",
// 		Func: cd,
// 	})

// 	// handle "load".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "load",
// 		Help: "load file into database",
// 		Func: load,
// 	})

// 	// handle "extract".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "extract",
// 		Help: "extract global",
// 		Func: extract,
// 	})

// 	// handle "walk".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "walk",
// 		Help: "walk global",
// 		Func: walk,
// 	})

// 	// handle "exec".
// 	shell.AddCmd(&ishell.Cmd{
// 		Name: "exec",
// 		Help: "exec statement",
// 		Func: exec,
// 	})

// 	// when started with "exit" as first argument, assume non-interactive execution
// 	if len(args) > 0 && args[1] == "exit" {
// 		shell.Process(args[1:]...)
// 	} else {
// 		// start shell
// 		shell.Run()
// 		// teardown
// 		shell.Close()
// 	}

// 	return nil
// }
