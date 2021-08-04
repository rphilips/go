package tools

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func IsUTF8(body []byte, lineno int) (err error) {
	if !utf8.Valid(body) {
		lines := bytes.SplitN(body, []byte("\n"), -1)
		for i, line := range lines {
			if !utf8.Valid(line) {
				err = fmt.Errorf("error line %d: No valid UTF-8 in `%s`", lineno+i, line)
				return
			}
		}
	}
	repl := rune(65533)
	if bytes.ContainsRune(body, repl) {
		lines := bytes.SplitN(body, []byte("\n"), -1)
		for i, line := range lines {
			if bytes.ContainsRune(line, repl) {
				err = fmt.Errorf("error line %d: No valid UTF-8 in `%s`", lineno+i, line)
				return
			}
		}
	}
	return nil
}

func LeftStrip(body []byte) (result []byte, extra int) {
	for i, b := range body {
		if unicode.IsSpace(rune(b)) {
			if b == '\n' {
				extra++
			}
			continue
		}
		result = body[i:]
		break
	}
	return
}

func LeftWord(body string) (before string, after string) {
	for i, r := range body {
		if unicode.IsLetter(r) {
			continue
		}
		before = body[:i]
		after = body[i:]
		break
	}
	return
}

func FirstAlfa(body string) (before string, after string) {
	for i, r := range body {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			before = body[:i]
			after = body[i:]
			break
		}
	}
	return
}

func ParseIsoDate(body string) (mytime time.Time, err error) {
	body = strings.TrimSpace(body)
	re := regexp.MustCompile(`[^0-9]+`)
	parts := re.Split(body, -1)
	if len(parts) == 3 {
		d := parts[0] + "-" + parts[1] + "-" + parts[2] + "T00:00:00Z"
		mytime, err = time.Parse(time.RFC3339, d)
		if err == nil {
			return
		}
	}

	err = fmt.Errorf("invalid date `%s`", body)
	return

}
