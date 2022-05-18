package action

import (
	"fmt"
	"io"
	"log"
	"strings"

	qfs "brocade.be/base/fs"
	qliner "github.com/peterh/liner"
)

func AWK(text string) []string {
	gloref, value := SplitRefValue(text)
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	var err error = nil
	setme := gloref + "=" + value
	awk := ""
	for {
		setme, err = setter.PromptWithSuggestion("    >> awk ", setme, -1)
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
		gloref, awk = SplitRefValue(setme)

		// h := []string{"set " + gloref + "=" + value}
		// if gloref2 != gloref {
		// 	h = append(h, "set "+gloref2+"="+value)
		// }
		// return h
		break
	}
	awk = strings.TrimSpace(awk)
	data, err := qfs.Fetch(awk)
	if err != nil {
		awk = string(data)
	}
	fmt.Println("AWK = [" + awk + "]")

	return nil
}
