package util

import (
	"strings"

	qliner "github.com/peterh/liner"
)

type Ask struct {
	Prompt  string
	Repeat  bool
	IsBoole bool
}

var Askit = map[string]Ask{
	"awk": {
		Prompt: "AWK command",
	},
	"backup": {
		Prompt: "Backup file",
	},
	"sed": {
		Prompt: "sed command",
	},
	"search": {
		Prompt: "Search for",
	},
	"replacement": {
		Prompt: "Replacement string",
	},
	"property": {
		Prompt: "Brocade property (naked|process|qtech|script|temp|web|webdav)",
	},
	"replace": {
		Prompt: "Replace with",
	},
	"files": {
		Prompt: "File/directory",
		Repeat: true,
	},
	"recurse": {
		Prompt:  "Recurse through directories ?",
		IsBoole: true,
	},
	"windows": {
		Prompt:  "Windows EOL convention ?",
		IsBoole: true,
	},
	"unix": {
		Prompt:  "Unix EOL convention ?",
		IsBoole: true,
	},
	"regexp": {
		Prompt:  "Regular expression ?",
		IsBoole: true,
	},
	"url": {
		Prompt:  "Show as URL ?",
		IsBoole: true,
	},
	"tolower": {
		Prompt:  "To lowercase ?",
		IsBoole: true,
	},
	"isfile": {
		Prompt:  "Is this a file ?",
		IsBoole: true,
	},
	"confirm": {
		Prompt:  "Ask for confirmation the first time ?",
		IsBoole: true,
	},
	"delete": {
		Prompt:  "Delete source ?",
		IsBoole: true,
	},
	"patterns": {
		Prompt: "File basename pattern",
		Repeat: true,
	},
	"utf8only": {
		Prompt:  "Restrict to UTF-8 files ?",
		IsBoole: true,
	},
	"ext": {
		Prompt: "Extension to basename",
	},
}

func AskArgs(codes []string) (result map[string]interface{}, aborted bool) {
	asker := qliner.NewLiner()
	asker.SetCtrlCAborts(true)
	defer asker.Close()

	result = make(map[string]interface{})

	for _, code := range codes {
		parts := strings.SplitN(code, ":", 3)
		parts = append(parts, "", "", "")
		code = strings.TrimSpace(parts[0])
		checks := strings.TrimSpace(parts[1])
		defa := strings.TrimSpace(parts[2])
		code := strings.TrimSpace(code)
		ask := Askit[code]
		repeat := ask.Repeat
		prompt := ask.Prompt + ": "
		if prompt == "" {
			continue
		}
		if !IsTrue(checks, result) {
			if repeat {
				result[code] = []string{}
			} else {
				if ask.IsBoole {
					result[code] = Yes(defa)
				} else {
					result[code] = defa
				}

			}
			continue
		}
		if repeat {
			give := make([]string, 0)
			for {
				giv, err := asker.PromptWithSuggestion(prompt, defa, -1)
				if err == qliner.ErrPromptAborted {
					return nil, true
				}
				giv = strings.TrimSuffix(giv, "\n")
				if giv == "" {
					result[code] = give
					break
				}
				give = append(give, giv)
			}
		} else {
			give, err := asker.PromptWithSuggestion(prompt, defa, -1)
			if err == qliner.ErrPromptAborted {
				return nil, true
			}
			if ask.IsBoole {
				result[code] = Yes(give)
			} else {
				result[code] = give
			}
		}
	}
	return result, false

}

func UnYes(b bool) string {
	if b {
		return "y"
	}
	return "n"
}

func IsTrue(checks string, result map[string]interface{}) bool {
	if checks == "" {
		return true
	}
	parts := strings.SplitN(checks, ",", -1)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		not := strings.HasPrefix(part, "!")
		part = strings.TrimLeft(part, "! ")
		if part == "" {
			continue
		}
		value, ok := result[part]
		if !ok {
			return not
		}
		switch v := value.(type) {
		case string:
			if not && v != "" {
				return false
			}
			if !not && v == "" {
				return false
			}
		case bool:
			if not && v {
				return false
			}
			if !not && !v {
				return false
			}
		case []string:
			if not && len(v) != 0 {
				return false
			}
			if !not && len(v) == 0 {
				return false
			}
		}
	}
	return true

}

func Confirm(prompt string) bool {
	asker := qliner.NewLiner()
	asker.SetCtrlCAborts(true)
	defer asker.Close()
	defa := "n"
	giv, err := asker.PromptWithSuggestion(prompt, defa, -1)
	if err == qliner.ErrPromptAborted {
		return false
	}
	giv = strings.TrimSuffix(giv, "\n")
	return Yes(giv)
}

func Yes(value string) bool {
	t := strings.ToLower(value)
	if t != "" && strings.Contains("jy1t", string(t[0])) {
		return true
	}
	return false
}

func AskArg(args []string, start int, rec bool, withpat bool, withutf8 bool, regexp bool) ([]string, bool, []string, bool, bool) {
	return nil, false, nil, false, false
}

// 	for {
// 		fmt.Print("File/directory: ")
// 		text, _ := reader.ReadString('\n')
// 		text = strings.TrimSpace(text)
// 		if text == "" {
// 			break
// 		}
// 		extra = append(extra, text)
// 	}
// 	if len(extra) == 0 {
// 		return nil, false, nil, false, false
// 	}
// 	recurse = false
// 	if askrecurse {
// 		fmt.Print("Recurse ? : <n>")
// 		text, _ := reader.ReadString('\n')
// 		text = strings.TrimSpace(text)
// 		if text == "" {
// 			text = "n"
// 		}
// 		if strings.ContainsAny(text, "jJyY1tT") {
// 			recurse = true
// 		}
// 	}

// 	if askpattern {
// 		for {
// 			fmt.Print("Pattern on basename: ")
// 			text, _ := reader.ReadString('\n')
// 			text = strings.TrimSpace(text)
// 			if text == "" {
// 				break
// 			}
// 			patterns = append(patterns, text)
// 			if text == "*" {
// 				break
// 			}
// 		}
// 	}

// 	utf8only = false
// 	if askutf8 {
// 		fmt.Print("UTF-8 only ? : <n>")
// 		text, _ := reader.ReadString('\n')
// 		text = strings.TrimSpace(text)
// 		if text == "" {
// 			text = "n"
// 		}
// 		if strings.ContainsAny(text, "jJyY1tT") {
// 			utf8only = true
// 		}
// 	}

// 	regexp = false
// 	if askutf8 {
// 		fmt.Print("Regexp ? : <n>")
// 		text, _ := reader.ReadString('\n')
// 		text = strings.TrimSpace(text)
// 		if text == "" {
// 			text = "n"
// 		}
// 		if strings.ContainsAny(text, "jJyY1tT") {
// 			regexp = true
// 		}
// 	}
// 	return extra, recurse, patterns, utf8only, regexp
// }
