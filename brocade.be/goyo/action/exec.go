package action

import (
	"io"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
	qliner "github.com/peterh/liner"
)

func Exec(text string) []string {
	if text != "" {
		err := qyottadb.Exec(text)
		if err != nil {
			qutil.Error(err)
			return nil
		}
		return []string{text}
	}
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	deflt := ""
	history := make([]string, 0)
	for {
		text, err := setter.PromptWithSuggestion("$ ", deflt, -1)
		if err == qliner.ErrPromptAborted {
			break
		}
		if err == io.EOF {
			break
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		ltext := strings.ToLower(text)
		if ltext == "bye" || ltext == "exit" || ltext == "h" || ltext == "halt" {
			break
		}
		err = qyottadb.Exec(text)
		deflt = ""
		if err != nil {
			qutil.Error(err)
			deflt = text
		} else {
			history = append(history, text)
		}
	}
	return history
}
