package cmd

import (
	"github.com/abiosoft/ishell/v2"
	"github.com/fatih/color"
	"lang.yottadb.com/go/yottadb"

	qmumps "brocade.be/goyo/lib/mumps"
)

func set(c *ishell.Context) {
	var gloref string
	var err error
	stop := false
	value := ""
	green := color.New(color.FgGreen).SprintFunc()
	if len(c.Args) != 0 {
		answer := c.Args[0]
		gloref, value, err = qmumps.EditGlobal(answer)
		if err != nil {
			Fgloref = gloref
			Fvalue = value
		}
	}
	for !stop {
		c.Println(green("ref=value (empty to quit):"))
		c.ShowPrompt(false)
		defa := ""
		if Fgloref != "" {
			defa = Fgloref + "=" + Fvalue
		}
		answer := c.ReadLineWithDefault(defa)
		if answer == "" {
			stop = true
			continue
		}

		gloref, value, err = qmumps.EditGlobal(answer)
		if err != nil {
			c.Println("?", err.Error())
			continue
		}

		Fgloref = gloref
		Fvalue = value
		gloref, subs, _ := qmumps.GlobalRef(Fgloref)
		Fgloref = gloref
		err = yottadb.SetValE(yottadb.NOTTP, nil, Fvalue, subs[0], subs[1:])
		if err != nil {
			c.Println(err.Error())
		}
	}
	c.ShowPrompt(true)
}
