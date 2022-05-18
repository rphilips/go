package action

type HelpData struct {
	Ref   string
	Short string
	Long  string
}

var IOcsvout string
var Actions = map[string]HelpData{
	"help": {
		Ref:   "",
		Short: "Help on goyo REPL",
	},
	"#": {
		Ref:   "",
		Short: "Start a comment",
	},
	"bye": {
		Ref:   "",
		Short: "End the REPL",
		Long: `Together with 'quit', 'exit', 'CTRL+D', 'CTRL+C', 'bye'
ends the REPL loop and returns to the shell`,
	},
	"cd": {
		Ref:   "",
		Short: "Changes the working directory",
		Long: `Changes the working directory:

	- without argument, changes to the Desktop or to the
	  home directory (depending on the availability of a desktop)
	- '~' is replaced in argument and directory is changed




		Together with 'quit', 'exit', 'CTRL+D', 'CTRL+C', 'bye'
ends the REPL loop and returns to the shell`,
	},
	"echo": {
		Ref:   "",
		Short: "Echoes the argument",
	},
	"exec": {
		Ref:   "",
		Short: "Executes a string in M",
		Long: `With an argument, this argument is executed in the M environment,
without argument, the REPL reads lines from stdin and exexutes them,
one by one`,
	},
	"exit": {Ref: "bye"},
	"extract": {
		Ref:   "",
		Short: "Extract an M global to a file",
		Long: `Extract an M global to a file.
Example:

    extract BCAT`,
	},

	"load": {
		Ref:   "",
		Short: "Load a global from a file",
		Long: `Load an M global to a file.
Example:

    load BCAT.zwr`,
	},
	"quit": {Ref: "bye"},

	"set": {
		Ref:   "",
		Short: "Set a variable (local or global)",
		Long: `Set a variable (local or global). Subscripts can be
valid M expressions`,
	},
	"walk": {
		Ref:   "",
		Short: "Navigate a variable (local or global)",
		Long: `Navigate a variable (local or global). Subscripts can be
valid M expressions.

During navigation one-key instructions can be given:

	- n: (down arrow) next subscript
	- p: (up arrow) previous subscript
	- l: (left arrow) loose a subscript level
	- r: (right arrow) extra subscript level
	- s: set to a new value
	- e: edit to initialise to a new reference
	- k: kill tree
	- /string: search forward (both on string and on regexp)
`,
	},
	"kill": {
		Ref:   "",
		Short: "Kills tree",
	},
	"killtree": {Ref: "kill"},
	"killnode": {
		Ref:   "",
		Short: "Kills treeNODE",
	},
	"get": {
		Ref:   "",
		Short: "Shows data referenced by the argument",
	},
	"zwr": {
		Ref:   "",
		Short: "Shows zwr data referenced by the argument",
	},
	"csv": {
		Ref:   "",
		Short: "Shows global data in CSV style",
	},
	"spec": {
		Ref:   "",
		Short: "Specify an option (spec csvout /library/tmp/my.csv)",
	},
}

func RunAction(key string, text string) []string {
	switch key {
	case "spec":
		return Spec(text)
	case "cd":
		return Cd(text)
	case "echo":
		return Echo(text)
	case "load":
		return Load(text)
	case "extract":
		return Extract(text)
	case "set":
		return Set(text)
	case "kill", "killtree":
		return Kill(text, true)
	case "killnode":
		return Kill(text, false)
	case "get":
		return Get(text)
	case "zwr":
		return ZWR(text)
	case "csv":
		return CSV(text, IOcsvout)
	case "walk":
		return walk(text)
	case "exec":
		return Exec(text)
	case "help":
		return Help(text)
	}
	return nil
}
