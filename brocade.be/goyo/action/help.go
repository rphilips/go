package action

import (
	"fmt"
	"sort"
	"strings"
)

func Help(text string) []string {
	text = strings.ToLower(text)
	action, ok := Actions[text]
	if !ok {
		text = ""
	}
	if text == "" {
		actions := make([]string, 0)
		for key := range Actions {
			actions = append(actions, key)
		}
		sort.Strings(actions)

		for _, verb := range actions {
			action := Actions[verb]
			doc := action.Short
			if doc == "" {
				ref := action.Ref
				action = Actions[ref]
				doc = action.Short
			}
			fmt.Printf("%-10v %s\n", verb, doc)

		}
	}

	if action.Ref != "" {
		ref := action.Ref
		action = Actions[ref]
	}
	doc := action.Long
	if doc == "" {
		doc = action.Short
	}
	fmt.Println(doc)
	return nil
}
