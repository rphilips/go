package completer

import (
	"strings"

	qliner "github.com/peterh/liner"
)

var actions = map[string]bool{
	"bye":      true,
	"cd":       true,
	"echo":     true,
	"exec":     true,
	"exit":     true,
	"extract":  true,
	"greet":    true,
	"load":     true,
	"quit":     true,
	"repl":     true,
	"set":      true,
	"walk":     true,
	"kill":     true,
	"killtree": true,
	"killnode": true,
	"get":      true,
	"zwr":      true,
	"def":      true,
	"defined":  true,
}

func SetCompleter(line *qliner.State) {
	fn := func(line string) (c []string) {
		for n := range actions {
			if strings.HasPrefix(n, strings.ToLower(line)) {
				c = append(c, n)
			}
		}
		return
	}
	line.SetCompleter(fn)
}

func IsAction(action string) bool {
	return actions[strings.ToLower(action)]
}
