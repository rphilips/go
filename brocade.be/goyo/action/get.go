package action

import (
	"fmt"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func Get(text string) []string {
	gloref, _ := SplitRefValue(text)
	value, err := qyottadb.G(gloref, false)
	if err != nil {
		qutil.Error(err)
		return nil
	}
	fmt.Println(value)
	gloref2 := qyottadb.N(gloref)
	h := []string{"get " + gloref}
	if gloref2 != gloref {
		h = append(h, "get "+gloref2)
	}
	return h
}
