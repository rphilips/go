package action

import (
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func Search(text string, needle string) string {
	if text == "" || needle == "" {
		return ""
	}
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	show := make(chan qyottadb.VarReport)
	go qyottadb.ZWR(gloref2, show, needle)
	report := <-show
	return report.Gloref
}
