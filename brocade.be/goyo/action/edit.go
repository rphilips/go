package action

import (
	"io"
	"log"
	"strings"

	qliner "github.com/peterh/liner"
)

func Edit(text string) string {
	gloref, _ := SplitRefValue(text)
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	var err error = nil
	setme := gloref
	for {
		setme, err = setter.PromptWithSuggestion("    >> edit ", setme, -1)
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
		gloref, _ = SplitRefValue(setme)
		return gloref
	}
	return ""
}
