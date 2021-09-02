package cmd

import (
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qmumps "brocade.be/goyo/lib/mumps"
	"github.com/abiosoft/ishell/v2"
)

func load(c *ishell.Context) {
	argums := make([]string, 0)
	if len(c.Args) != 0 {
		argums = append(argums, "LOAD")
		for _, arg := range c.Args {
			if strings.HasPrefix(arg, "-") {
				argums = append(argums, arg)
				continue
			}
			if qfs.IsFile(arg) {
				argums = append(argums, arg)
				continue
			}
			matches, err := filepath.Glob(arg)
			if err != nil {
				argums = append(argums, arg)
				continue
			}
			argums = append(argums, matches...)
		}
	}
	if len(argums) == 0 {
		help := `
	mupip load
		[-BE[GIN]=integer -E[ND]=integer
		-FI[LLFACTOR]=integer
		-FO[RMAT]={GO|B[INARY]|Z[WR]]}
		-I[GNORECHSET]
		-O[NERROR]={STOP|PROCEED|INTERACTIVE}
		-S[TDIN]] file-name`
		c.Println(qmumps.Info(help))
	} else {

		stdout, stderr, err := qmumps.MUPIP(argums, "")
		f := qmumps.Error
		if err == nil {
			f = qmumps.Info
		}
		if strings.TrimSpace(stderr) != "" {
			c.Println(f(stderr))
		}
		if strings.TrimSpace(stdout) != "" {
			c.Println(f(stdout))
		}
		if err != nil {
			c.Println(qmumps.Error(err.Error()))
		}
	}
	c.ShowPrompt(true)
}
