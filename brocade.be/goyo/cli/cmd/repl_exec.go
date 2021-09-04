package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell/v2"
	"lang.yottadb.com/go/yottadb"

	qenv "brocade.be/goyo/lib/env"
	qmumps "brocade.be/goyo/lib/mumps"
)

func exec(c *ishell.Context) {
	answer := ""
	if answer == "" {
		answer = Fexec
	}
	c.Println(qmumps.Ask("exec (empty to quit):"))
	exec := handleexec(c, answer)

	if exec == "" {
		c.ShowPrompt(true)
		return
	}
	fmtable := "/home/rphilips/.yottadb/ydbaccess.ci"
	envvarSave := make(map[string]string)
	qenv.SaveEnvvars(&envvarSave, "ydb_ci", "ydb_routines")
	os.Setenv("ydb_ci", fmtable)
	out := strings.Repeat(" ", 1024*64)
	_, err := yottadb.CallMT(yottadb.NOTTP, nil, 0, "xecute", exec, &out)
	fmt.Println("ERROR:", out)
	qenv.RestoreEnvvars(&envvarSave, "ydb_ci", "ydb_routines")
	c.ShowPrompt(true)
	if nil != err {
		panic(fmt.Sprintf("CallMT() call failed: %s", err))
	}
}

func handleexec(c *ishell.Context, exec string) string {
	c.ShowPrompt(false)
	exec = c.ReadLineWithDefault(exec)
	exec = strings.TrimSpace(exec)
	return exec
}
