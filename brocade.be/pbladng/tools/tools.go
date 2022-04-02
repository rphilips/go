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

func NumberSplit(body string, money bool, start int) (before string, number string, after string) {
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
		return NumberSplit(unescape(body), money, k+l+1)
	}
	before = body[:k]
	for i, r := range body[k:] {
		if r > 47 && r < 58 {
			continue
		}
		if money && r == 44 {
			money = false
			continue
		}
		number = body[k : k+i]
		after = body[k+i:]
		break
	}
	if number == "" {
		number = body[k:]
	}
	if strings.HasSuffix(number, ",") {
		number = strings.TrimSuffix(number, ",")
		after = "," + after
	}
	before = unescape(before)
	after = unescape(after)
	return
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

// Euro transformation
func Euro(body string) string {
	body = strings.ReplaceAll(body, "€", " EUR ")
	body = escape(body)
	start := 0
	for {
		before, number, after := NumberSplit(body, true, start)
		if number == "" {
			break
		}
		punct, word, rest := FirstAlfa(after)
		punct = strings.ReplaceAll(punct, " ", "")
		punct = strings.ReplaceAll(punct, "\t", "")
		if punct != "" {
			start = len(before) + len(number)
			continue
		}
		word = strings.ToUpper(word)
		if word == "EUR" || word == "EURO" {
			body = before + "{" + number + " EUR}" + rest
			start = len(before) + len(number) + len("{ EUR}")
			continue
		}
		start = len(before) + len(number)
	}
	return unescape(body)
}

// Display phone

func showphone(x string) string {
	zones := []string{
		"02", "03", "09", "010",
		"011", "012", "013", "014", "015",
		"016", "019", "050", "051", "052",
		"053", "054", "055", "056", "057",
		"058", "059", "060", "061", "063",
		"064", "065", "067", "069", "071",
		"080", "081", "082", "083", "085",
		"086", "087", "089",
	}
	rex := regexp.MustCompile(`[^0-9]`)
	x = rex.ReplaceAllString(x, "")
	zone := ""
	for _, p := range zones {
		if strings.HasPrefix(x, p) {
			zone = p
			break
		}
	}
	if zone == "" {
		zone = x[:4]
	}
	x = x[len(zone):]
	if len(x) == 6 {
		x = x[:2] + " " + x[2:4] + " " + x[4:]
	} else {
		x = x[:3] + " " + x[3:5] + " " + x[5:]
	}
	return "(" + zone + ") " + x
}

// Phone transformation
func Phone(body string) string {

	// 09 385 62 03
	// 0475 812 419
	rex := regexp.MustCompile(`0[1-9]([().]? *[0-9]){7,8}`)
	body = escape(body)

	phones := rex.FindAllStringIndex(body, -1)
	if len(phones) == 0 {
		return unescape(body)
	}
	result := ""

	for i, phone := range phones {
		if i == 0 {
			result = body[:phone[0]]
		} else {
			result += body[phones[i-1][1]:phone[0]]
		}
		if strings.Count(result, "{") != strings.Count(result, "}") {
			result += body[phone[0]:phone[1]]
			continue
		}
		result += "{" + showphone(body[phone[0]:phone[1]]) + "}"
	}
	phone := phones[len(phones)-1]
	result += body[phone[1]:]

	return unescape(result)
}

// func ApplyDates(body string, lastdate string, start int) (string, string) {
// 	before, number, after := NumberSplit(body, false, start)
// 	if number == "" {
// 		return body, lastdate
// 	}

// 	// patterns

// 	// 17 juli 2021
// 	// 17 -07 -2021
// 	// 17 juli '21
// 	// 17-07-21

// 	// datum voor de punctuatie
// 	punct, word, rest := FirstAlfa(after)
// 	if rest == "" || strings.ContainsAny(punct, ".?!") {
// 		done, todo, date := beforeDate(before, number, punct, word, rest)
// 		if date > lastdate {
// 			lastdate = date
// 		}
// 		return ApplyDates(done+todo, lastdate, len(done))
// 	}
// 	// punctuatie niet compleet
// 	punct = escape(punct)
// 	if strings.Count(punct, "{") != strings.Count(punct, "}") {
// 		rest = escape(rest)
// 		k := strings.Index(rest, "}")
// 		if k == -1 {
// 			return body, lastdate
// 		}
// 		return ApplyDates(body, lastdate, start+len(punct)+k+1)
// 	}
// 	// datum wordt achteraf bepaald
// 	// done, todo, date := afterDate(before, number, punct, word, rest)
// 	// if date > lastdate {
// 	// 	lastdate = date
// 	// }
// 	return ApplyDates(done+todo, lastdate, len(done))
// }

// func beforeDate(before, number, punct, word, rest string) (done, todo, datum string) {
// 	bef, wrd, aft := LastAlfa(before)
// 	if wrd == "" {

// 	}

// }

// 17 juli
// 17 juli 2021
// 17 juli '21
// 17-07
// 17-7-2021
// 17-7-21
// juli, 17
// juli, 17, 2021

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
	for _, r := range `|*_` {
		if strings.Count(body, string(r))%2 != 0 {
			return fmt.Errorf("uneven number of `%s`", string(r))
		}
	}
	for _, r := range `|*_` {
		ch := string(r)
		ech := regexp.QuoteMeta(ch)
		rex, _ := regexp.Compile(ech + "[^" + ch + "]*" + ech)
		im := rex.ReplaceAllString(body, "")
		set := strings.Replace(`|*_`, ch, "", -1)
		for _, s := range set {
			if strings.Count(im, string(s))%2 != 0 {
				return fmt.Errorf("uneven number of `%s` in substring starting with `%s`", string(s), ch)
			}
		}
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
