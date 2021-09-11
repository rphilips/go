package cmd

import (
	"fmt"
	"io"

	qaction "brocade.be/goyo/action"
	qutil "brocade.be/goyo/lib/util"
	qprompt "github.com/manifoldco/promptui"
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

func finish() {
	fmt.Println("Bye!")
	// rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
	// rawModeOff.Stdin = os.Stdin
	// _ = rawModeOff.Run()
	// rawModeOff.Wait()
	return
}

var Fgloref string
var Fexec string
var Fvalue string
var Fexit bool

func repl(cmd *cobra.Command, args []string) error {
	defer finish()
	validate := func(input string) error {
		return nil
	}
	Fexit = false
	keyword := ""
	text := ""
	fmt.Println("Please use `exit` to exit this program.")
	prompt := qprompt.Prompt{
		Label:    Defaults["prompt-repl"],
		Validate: validate,
		Default:  "",
	}

	for !Fexit {
		todo, err := prompt.Run()

		if err != nil {
			Fexit = true
			continue
		}
		if err == io.EOF {
			err = nil
			todo = "quit"
		}
		if keyword == "" {
			keyword, text = qutil.KeyText(todo)
			if keyword == "" {
				continue
			}
		}
		switch keyword {
		case "exit":
			keyword = ""
			Fexit = true

		case "about":
			keyword = ""
			fmt.Println(AboutText())
		case "cd":
			keyword = ""
			qaction.Cd(text)
		case "set":
			keyword = ""
			qaction.Set(text)
		default:
			fmt.Printf("Unknown input: [%s] [%s]\n", keyword, text)

		}
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
