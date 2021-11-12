package action

import (
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func ZWR(text string) []string {
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	qyottadb.Exec("zwr " + gloref2)
	h := []string{"zwr " + gloref}
	if gloref2 != gloref {
		h = append(h, "zwr "+gloref2)
	}
	return h
}
