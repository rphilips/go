package action

import (
	"io"
	"log"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
	qliner "github.com/peterh/liner"
)

func Set(text string) []string {
	gloref, value := SplitRefValue(text)
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	var err error = nil
	setme := gloref + "=" + value
	for {
		setme, err = setter.PromptWithSuggestion("    >> set ", setme, -1)
		if err == qliner.ErrPromptAborted {
			break
		}
		if strings.TrimSpace(setme) == "" {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print("Error reading line: ", err)
			continue
		}
		gloref, value = SplitRefValue(setme)
		err = qyottadb.Set(gloref, value)
		if err != nil {
			qutil.Error(err)
			continue
		}
		gloref2 := qyottadb.N(gloref)
		h := []string{"set " + gloref + "=" + value}
		if gloref2 != gloref {
			h = append(h, "set "+gloref2+"="+value)
		}
		return h
	}
	return nil
}

func SplitRefValue(text string) (ref string, value string) {
	start := 0
	for {
		k := strings.Index(text[start:], "=")
		if k < 0 {
			ref = qyottadb.UnQS(qyottadb.QS(text))
			value, _ = qyottadb.G(ref, false)
			break
		}
		k = start + k
		if strings.Count(text[:k], `"`)%2 == 1 {
			start = k + 1
			continue
		}

		value = text[k+1:]
		ref = qyottadb.UnQS(qyottadb.QS(text[:k]))
		break
	}
	return
}
