package cmd

import (
	"strings"

	qmumps "brocade.be/goyo/lib/mumps"
	"github.com/abiosoft/ishell/v2"
)

func extract(c *ishell.Context) {
	argums := make([]string, 0)
	glo := ""
	fmt := ".zwr"
	if len(c.Args) != 0 {
		argums = append(argums, "EXTRACT")
		for _, arg := range c.Args {
			if strings.HasPrefix(strings.ToUpper(arg), "-FO") && strings.ContainsRune(arg, '=') {
				x := strings.SplitN(arg, "=", 2)[1]
				if x != "" {
					fmt = "." + strings.ToLower(x)
				}
			}
			if strings.HasPrefix(arg, "-") {
				argums = append(argums, arg)
				continue
			}
			arg = strings.TrimPrefix(arg, "^")
			argums = append(argums, "-SE="+arg)
			glo = arg
			break

			// if qfs.IsFile(arg) {
			// 	argums = append(argums, arg)
			// 	continue
			// }
			// matches, err := filepath.Glob(arg)
			// if err != nil {
			// 	argums = append(argums, arg)
			// 	continue
			// }
			// argums = append(argums, matches...)
		}
	}
	if len(argums) == 0 {
		help := `
mupip extract
	[-FO[RMAT]={GO|B[INARY]|Z[WR]}
	-FR[EEZE]
	-LA[BEL]=text
	-[NO]L[OG]
	-[NO]NULL_IV
	-R[EGION]=region-list
	-SE[LECT]=global-name-list]
	]
	{-ST[DOUT]|file-name}`

		c.Println(qmumps.Info(help))
	} else {
		if glo != "" {
			argums = append(argums, glo+fmt)
		}

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
