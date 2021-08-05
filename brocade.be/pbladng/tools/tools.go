package tools

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func IsUTF8(body []byte, lineno int) (err error) {
	if len(body) == 0 {
		return nil
	}
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

// LeftTrim removes whitespace at the beginning and counts the removed \n
func LeftTrim(body string) (result string, extra int) {
	if body == "" {
		return
	}
	for i, r := range body {
		if unicode.IsSpace(r) {
			if r == '\n' {
				extra++
			}
			continue
		}
		result = body[i:]
		break
	}
	return
}

// LeftWord splits a string in a word(letters only) and the rest
func LeftWord(body string) (word string, after string) {
	word = body
	for i, r := range body {
		if unicode.IsLetter(r) {
			continue
		}
		word = body[:i]
		after = body[i:]
		break
	}
	return
}

// FirstAlfa splits a string until the first alfa (nummer, digit)
func FirstAlfa(body string) (before string, word string, rest string) {
	inword := -1
	for i, r := range body {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if inword < 0 {
				before = body[:i]
				inword = i
				word = body[i:]
			}
			continue
		}
		if inword != -1 {
			word = body[inword:i]
			rest = body[i:]
			break
		}
	}
	return
}

// LastAlfa splits a string until the last alfa (nummer, digit)
func LastAlfa(body string) (before string, word string, after string) {
	if body == "" {
		return
	}
	inword := -1
	wpos := -1
	last := len(body)
	for last > 0 {
		r, size := utf8.DecodeLastRuneInString(body[:last])
		last -= size
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if inword < 0 {
				after = body[last+size:]
				inword = last + size
				word = body[:last+size]
			}
			wpos = last
			continue
		}
		if inword != -1 {
			word = body[wpos:inword]
			before = body[:wpos]
			break
		}

	}
	return
}

// ParseIsoDate make a time.Tine and error of string
func ParseIsoDate(body string) (mytime time.Time, err error) {
	body = strings.TrimSpace(body)
	re := regexp.MustCompile(`[^0-9]+`)
	parts := re.Split(body, -1)
	if len(parts) == 3 {
		if len(parts[0]) < 4 {
			parts[0] = "0000"[:4-len(parts[0])] + parts[0]
		}
		if len(parts[1]) < 2 {
			parts[1] = "0000"[:2-len(parts[1])] + parts[1]
		}
		if len(parts[2]) < 2 {
			parts[2] = "0000"[:2-len(parts[2])] + parts[2]
		}

		d := parts[0] + "-" + parts[1] + "-" + parts[2] + "T00:00:00Z"
		mytime, err = time.Parse(time.RFC3339, d)
		if err == nil {
			return
		}
	}
	err = fmt.Errorf("invalid date `%s`", body)
	return

}

// NumberSplit

func NumberSplit(body string, start int) (before string, number string, after string) {
	if len(body) < start {
		return body, "", ""
	}
	k := strings.IndexAny(body[start:], "1234567890")
	if k == -1 {
		return body, "", ""
	}
	k += start
	body = escape(body)
	x := body[:k]
	if strings.Count(x, "{") != strings.Count(x, "}") {
		l := strings.Index(body[k:], "}")
		if l == -1 {
			return unescape(body), "", ""
		}
		return NumberSplit(unescape(body), k+l+1)
	}
	before = body[:k]
	for i, r := range body[k:] {
		if r > 47 && r < 58 {
			continue
		}
		number = body[k : k+i]
		after = body[k+i:]
		break
	}
	if number == "" {
		number = body[k:]
	}
	before = unescape(before)
	after = unescape(after)
	return
}

func NumberSplitter(body string) []string {
	parts := make([]string, 0)
	before, number, after := NumberSplit(body, 0)
	parts = append(parts, before)
	if number == "" {
		return parts
	}
	parts = append(parts, number)
	parts = append(parts, NumberSplitter(after)...)
	return parts
}

func Euro(body string) string {

	return ""
}

func escape(body string) string {
	if len(body) < 2 {
		return body
	}
	set := `\|*_{}`
	if !strings.ContainsAny(body, set) {
		return body
	}
	for i, r := range set {
		if !strings.ContainsRune(body, r) {
			continue
		}
		rs := string([]byte{byte(i), byte(i)})
		body = strings.ReplaceAll(body, `\`+string(r), rs)
	}
	return body
}

func unescape(body string) string {
	set := "\x00\x01\x02\x03\x04\x05"
	oset := `\|*_{}`
	if !strings.ContainsAny(body, set) {
		return body
	}
	for i, r := range set {
		if !strings.ContainsRune(body, r) {
			continue
		}
		rs := string([]byte{byte(r), byte(r)})
		body = strings.ReplaceAll(body, rs, `\`+string(oset[i]))
	}
	return body
}

func euro(body string) string {
	body = escape(body)
	rex := regexp.MustCompile(`[Ee][Uu][Rr][Oo]?`)
	rex.Longest()
	parts := rex.FindAllStringIndex(body, -1)
	if parts == nil {
		return unescape(body)
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
	return unescape(keep)
}

// func ApplyDates(body string, lastdate string, start int) (result string, lastdate string) {
// 	parts := NumberSplitter(body)

// 	before, number, after := NumberSplit(body, start)
// 	if number == "" {
// 		return body, lastdate
// 	}
// 	punct, rest := FirstAlfa(after)
// 	punct = escape(punct)
// 	if strings.Count(punct, "{") != strings.Count(punct, "}") {
// 		rest = escape(rest)
// 		k := strings.Index(rest, "}")
// 		if k == -1 {
// 			return body, lastdate
// 		}
// 		return ApplyDates(body, lastdate, start+len(punct)+k+1)
// 	}
// 	// 17 juli
// 	// 17 juli 2021
// 	// 17 juli '21
// 	// 17-07
// 	// 17-7-2021
// 	// 17-7-21
// 	// juli, 17
// 	// juli, 17, 2021

// 	if rest != "" && !strings.ContainsAny(punct,".?!") {
// 		word := LeftWord(rest)
// 		month := Month(word)
// 		if

// 	}

// 	return

// }

func Curly(body string) error {
	body = escape(body)
	rex := regexp.MustCompile("{[^{}]*}")
	y := body
	for {
		x := rex.ReplaceAllString(y, "")
		if x == y {
			break
		}
		y = x
	}
	if strings.ContainsAny(y, "{}") {
		return errors.New("problem with curly braces")
	}
	return nil
}

func Month(word string) string {
	type month struct {
		read  string
		write string
	}
	months := []month{
		{
			"jan",
			"januari",
		},
		{
			"feb",
			"februari",
		},
		{
			"fev",
			"februari",
		},
		{
			"maart",
			"maart",
		},
		{
			"mrt",
			"maart",
		},
		{
			"maa",
			"maart",
		},
		{
			"apr",
			"april",
		},
		{
			"mei",
			"mei",
		},
		{
			"jun",
			"juni",
		},
		{
			"jul",
			"juli",
		},
		{
			"july",
			"juli",
		},
		{
			"aug",
			"augustus",
		},
		{
			"sep",
			"september",
		},
		{
			"oct",
			"oktober",
		},
		{
			"october",
			"oktober",
		},
		{
			"nov",
			"november",
		},
		{
			"dec",
			"december",
		},
	}

	word = strings.ToLower(word)

	for _, m := range months {
		if word == m.read {
			return m.write
		}
		if word == m.write {
			return m.write
		}
		if strings.HasPrefix(word, m.write) && strings.HasPrefix(m.write, word) {
			return m.write
		}
	}
	return ""
}
