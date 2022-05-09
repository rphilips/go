package action

import (
	"io"
	"strings"

	qhistory "brocade.be/goyo/lib/history"
	qliner "github.com/peterh/liner"
)

func Spec(text string) []string {
	deflt := text
	if text != "" {
		k := strings.IndexAny(text, " \t")
		prefix := text[:k]
		post := ""
		if prefix == "csvout" {
			post = strings.TrimSpace(text[k:])
		}
		if post != "" {
			IOcsvout = post
			deflt = ""
		}
		return []string{text}
	}
	setter := qliner.NewLiner()
	setter.SetCtrlCAborts(true)
	defer setter.Close()
	qhistory.LoadHistory(setter, "spec")
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
		k := strings.IndexAny(text, " \t")
		prefix := text[:k]
		post := ""
		if prefix == "csvout" {
			post = strings.TrimSpace(text[k:])
		}
		if post != "" {
			IOcsvout = post
			deflt = ""
		}
	}
	qhistory.SaveHistory(setter, "spec")
	setter.Close()
	return nil
}
