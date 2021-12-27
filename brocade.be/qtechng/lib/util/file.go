package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func AskArg(args []string, index int, askrecurse bool, askpattern bool, askutf8 bool, askregexp bool) (extra []string, recurse bool, patterns []string, utf8only bool, regexp bool) {
	reader := bufio.NewReader(os.Stdin)
	if len(args) != index {
		return nil, false, nil, false, false
	}

	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return nil, false, nil, false, false
	}

	for {
		fmt.Print("File/directory: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			break
		}
		extra = append(extra, text)
	}
	if len(extra) == 0 {
		return nil, false, nil, false, false
	}
	recurse = false
	if askrecurse {
		fmt.Print("Recurse ? : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			recurse = true
		}
	}

	if askpattern {
		for {
			fmt.Print("Pattern on basename: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if text == "" {
				break
			}
			patterns = append(patterns, text)
			if text == "*" {
				break
			}
		}
	}

	utf8only = false
	if askutf8 {
		fmt.Print("UTF-8 only ? : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			utf8only = true
		}
	}

	regexp = false
	if askutf8 {
		fmt.Print("Regexp ? : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			regexp = true
		}
	}
	return extra, recurse, patterns, utf8only, regexp
}
