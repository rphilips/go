package action

import (
	"io"
	"log"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
	qliner "github.com/peterh/liner"
)

func Kill(text string, tree bool) []string {
	gloref, value := SplitRefValue(text)
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	var err error = nil
	killme := gloref
	way := "killtree"
	if !tree {
		way = "killnode"
	}
	for {
		killme, err = setter.PromptWithSuggestion("    >> "+way+" ", killme, -1)
		if err == qliner.ErrPromptAborted {
			break
		}
		if strings.TrimSpace(killme) == "" {
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print("Error reading line: ", err)
			continue
		}

		err = qyottadb.Kill(gloref, tree)
		if err != nil {
			qutil.Error(err)
			continue
		}
		gloref2 := qyottadb.N(gloref)
		h := []string{way + " " + gloref}
		if gloref2 != gloref {
			h = append(h, way+" "+gloref2+"="+value)
		}
		return h
	}
	return nil
}
