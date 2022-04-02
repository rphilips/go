package action

import (
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func Search(text string, needle string, forward bool) string {
	if text == "" || needle == "" {
		return ""
	}
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	show := make(chan qyottadb.VarReport)
	go qyottadb.ZWR(gloref2, show, needle, forward)
	report := <-show
	return report.Gloref
}
