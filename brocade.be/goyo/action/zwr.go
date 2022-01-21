package action

import (
	"fmt"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func ZWR(text string) []string {
	if text == "" {
		return nil
	}
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	show := make(chan qyottadb.VarReport, 100)
	go qyottadb.ZWR(gloref2, show, "", true)
	for report := range show {
		if report.Err != nil {
			qutil.Error(report.Err)
			break
		}
		if report.Gloref == "" {
			break
		}
		fmt.Println(report.Gloref + "=" + report.Value)
	}
	h := []string{"zwr " + gloref}
	if gloref2 != gloref {
		h = append(h, "zwr "+gloref2)
	}
	return h
}
