package cmd

import (
	"os"
	"path/filepath"

	qfs "brocade.be/base/fs"
	qmumps "brocade.be/goyo/lib/mumps"
	"github.com/abiosoft/ishell/v2"
)

func cd(c *ishell.Context) {
	home, _ := os.UserHomeDir()
	argums := make([]string, 0)
	if len(c.Args) == 0 {
		desktop := filepath.Join(home, "Desktop")
		if qfs.IsDir(desktop) {
			argums = append(argums, desktop)
		} else {
			argums = append(argums, home)
		}
	} else {
		if c.Args[0] == "~" {
			argums = append(argums, home)
		} else {
			d, e := qfs.AbsPath(c.Args[0])
			if e != nil {
				d = c.Args[0]
			}
			argums = append(argums, d)
		}
	}
	err := os.Chdir(argums[0])
	if err == nil {
		cwd, _ := os.Getwd()
		c.Println(qmumps.Info(cwd))
	} else {
		c.Println(qmumps.Error(err.Error()))
	}

	c.ShowPrompt(true)
}
