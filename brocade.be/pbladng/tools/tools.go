package tools

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
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
	i := -1
	for _, b := range body {
		i++
		if unicode.IsSpace(rune(b)) {
			if b == '\n' {
				extra++
			}
			continue
		}
		break
	}
	if i != -1 {
		result = body[i:]
	}
	return
}

func LeftWord(body string) (before string, after string) {
	i := -1
	for _, r := range body {
		i++
		if unicode.IsLetter(r) {
			continue
		}
		break
	}
	if i != -1 {
		before = body[:i]
		after = body[i:]
	}
	return
}

func FirstAlfa(body string) (before string, after string) {
	i := -1
	for i, r := range body {
		i++
		if unicode.IsLetter(r) || unicode.IsDigit(r) {

			break
		}
	}
	if i != -1 {
		before = body[:i]
		after = body[i:]
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

func NumberSplit(body string) (before string, number int, after string) {
	k := strings.IndexAny(body, "1234567890")
	if k == -1 {
		return body, -1, ""
	}
	if k > 0 {
		before = body[:k]
	}

	number = -1
	for i, r := range body[k:] {
		if r > 47 && r < 58 {
			continue
		}
		number, _ = strconv.Atoi(body[k : k+i])
		after = body[k+i:]
		break
	}
	if number == -1 {
		number, _ = strconv.Atoi(body[k:])
	}
	return
}

func Euro(body string) string {

	return ""
}

func euro(body string) string {
	body = strings.ReplaceAll(body, `\{`, "\x01")
	body = strings.ReplaceAll(body, `\}`, "\x02")
	rex := regexp.MustCompile(`[Ee][Uu][Rr][Oo]?`)
	rex.Longest()
	parts := rex.FindAllStringIndex(body, -1)
	if parts == nil {
		return body
	}
	keep := ""
	last := 0
	for _, duo := range parts {
		k1 := duo[0]
		k2 := duo[1]
		if k1 > 0 {
			keep += body[last:k1]
		}
		last = k2
		if strings.Count(keep, "{") != strings.Count(keep, "}") {
			keep += body[k1:k2]
			continue
		}
		if keep != "" {
			r, _ := utf8.DecodeLastRuneInString(keep)
			if unicode.IsLetter(r) {
				keep += body[k1:k2]
				continue
			}
		}
		rest := body[k2:]
		if rest != "" {
			r, _ := utf8.DecodeRuneInString(rest)
			if unicode.IsLetter(r) {
				keep += body[k1:k2]
				continue
			}
		}
		keep += " EUR "
	}
	keep += body[last:]
	keep = regexp.MustCompile(` +EUR +`).ReplaceAllString(keep, " EUR ")
	keep = strings.ReplaceAll(keep, "\x01", `\{`)
	return strings.ReplaceAll(keep, "\x02", `\}`)
}
