package action

import (
	"io"
	"log"
	"strings"

	qliner "github.com/peterh/liner"
)

func Ask(prompt string, text string) string {
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	var err error = nil
	setme := text
	for {
		setme, err = setter.PromptWithSuggestion("    >> "+prompt, setme, -1)
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
		return setme
	}
	return ""
}
