package cmd

import (
	"strconv"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"lang.yottadb.com/go/yottadb"

	qmumps "brocade.be/goyo/lib/mumps"
)

func set(c *ishell.Context) {
	var gloref string
	var err error
	value := ""
	answer := ""
	if len(c.RawArgs) > 1 {
		answer = strings.Join(c.RawArgs[1:], " ")
		value, err = qmumps.GlobalValue(answer)
		if err == nil {
			answer += "=" + value
		}
	}

	if answer == "" {
		answer = Fgloref
	}
	c.Println(qmumps.Ask("ref=value (empty to quit):"))
	gloref, value, d := handleset(c, answer)
	c.ShowPrompt(true)
	if gloref != "" {
		Fgloref = gloref
	}
	c.Println(qmumps.Info(gloref) + "=" + qmumps.Info(value) + " " + qmumps.Error("$D()="+strconv.Itoa(d)))

}

func handleset(c *ishell.Context, gloref string) (string, string, int) {
	value, err := qmumps.GlobalValue(gloref)
	if err == nil {
		gloref += "=" + value
	}
	answer := gloref
	stop := false
	var subs []string
	c.ShowPrompt(false)
	for !stop {
		answer = c.ReadLineWithDefault(answer)
		if answer == "" {
			stop = true
			continue
		}
		gloref, value, err = qmumps.EditGlobal(answer)
		if err != nil {
			c.Println("?", qmumps.Error(err.Error()))
			continue
		}

		gloref, subs, _ = qmumps.GlobalRef(gloref)
		err = yottadb.SetValE(yottadb.NOTTP, nil, value, subs[0], subs[1:])
		if err != nil {
			c.Println(qmumps.Error(err.Error()))
			continue
		}
		stop = true
	}
	value, err = qmumps.GlobalValue(gloref)
	if err != nil {
		return gloref, value, -1
	}
	return gloref, value, qmumps.GlobalDefined(gloref)
}
