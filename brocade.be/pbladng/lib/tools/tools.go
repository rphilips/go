package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Line struct {
	L  string
	NR int
}

func WSLines(mylist []Line) (result []Line) {
	found := false
	prev := false
	last := -1
	for _, line := range mylist {
		s := strings.TrimSpace(line.L)
		if !found && s == "" {
			continue
		}
		found = true
		if prev && s == "" {
			continue
		}
		result = append(result, line)
		prev = s == ""
		if !prev {
			last = len(result)
		}
	}
	if last != -1 {
		result = result[:last]
	}
	return
}

func J(s any) string {
	js, _ := json.Marshal(s)
	return string(js)
}

func Normalize(s string) string {
	s = Phone(Latin1(s))
	s = Euro(s)
	s = FixColon(s)
	s = FixSpaceRune(s)
	re := regexp.MustCompile("  +")
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func TestLine(line Line) error {
	x := strings.ReplaceAll(line.L, "\\\\", "")
	for _, r := range []string{"|", "*", "_"} {
		x = strings.ReplaceAll(x, "\\"+r, "")
		nr := strings.Count(x, r)
		if nr%2 != 0 {
			return Error("line-unbalanced", line.NR, "unbalanced `"+r+"`")
		}
	}
	return nil
}

func FixSpaceRune(s string) string {
	if !strings.ContainsAny(s, "*_|") {
		return s
	}
	s = escape(s)
	for _, ch := range []string{"*", "_", "|"} {
		if !strings.Contains(s, ch) {
			continue
		}
		parts := strings.SplitN(s, ch, -1)
		for i, part := range parts {
			if len(parts) == (i + 1) {
				continue
			}
			switch i % 2 {
			case 0:
				x := strings.TrimLeft(parts[i+1], " \t\n\r")
				d := len(parts[i+1]) - len(x)
				if d == 0 {
					continue
				}
				parts[i] += parts[i+1][:d]
				parts[i+1] = x
			default:
				x := strings.TrimRight(part, " \t\n\r")
				d := len(part) - len(x)
				if d == 0 {
					continue
				}
				k := len(x)
				parts[i+1] = part[k:] + parts[i+1]
				parts[i] = x
			}
		}
		s = strings.Join(parts, ch)
		s = strings.ReplaceAll(s, ch+ch, "")
	}
	return unescape(s)
}

func IsUTF8(body []byte) (lines []string, err error) {
	if len(body) == 0 {
		return
	}
	repl := rune(65533)
	blines := bytes.SplitN(body, []byte("\n"), -1)
	lines = make([]string, len(blines))
	for i, bline := range blines {
		if !utf8.Valid(bline) {
			lines = nil
			err = Error("utf8-noutf8", i+1, "No valid UTF-8 in line")
			return
		}
		if bytes.ContainsRune(bline, repl) {
			lines = nil
			err = Error("utf8-repl", i+1, "Replacement character in line")
			return
		}
		lines[i] = string(bline)
	}
	return
}

func HeaderString(s string) string {
	k := strings.Index(s, "[")
	if k != -1 {
		s = strings.TrimSpace(s[:k])
	}
	s = Normalize(s)
	return strings.ToUpper(s)
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
	start := 0
	for {
		before, number, after := NumberSplit(body, true, start)
		if number == "" {
			break
		}
		after = strings.ReplaceAll(after, "\u00A0", " ")
		punct, word, rest := FirstAlfa(after)
		punct = strings.ReplaceAll(punct, " ", "")
		punct = strings.ReplaceAll(punct, "\t", "")
		if punct != "" {
			start = len(before) + len(number)
			continue
		}
		word = strings.ToUpper(word)
		if word == "EUR" || word == "EURO" {
			body = before + number + "\u00A0EUR" + rest
			start = len(before) + len(number) + len(" EUR")
			continue
		}
		start = len(before) + len(number)
	}
	return body
}

// Display phone

func showphone(x string) string {
	zones := []string{
		"090x", "080x", "02", "03", "09", "010",
		"011", "012", "013", "014", "015",
		"016", "019", "050", "051", "052",
		"053", "054", "055", "056", "057",
		"058", "059", "060", "061", "063",
		"064", "065", "067", "069", "071", "078",
		"080", "081", "082", "083", "085",
		"086", "087", "089",
	}
	rex := regexp.MustCompile(`[^0-9]`)
	x = rex.ReplaceAllString(x, "")
	zone := ""
	for _, z := range zones {
		p := z
		if strings.HasSuffix(p, "x") {
			p = strings.ReplaceAll(z, "x", "")
		}
		if strings.HasPrefix(x, p) {
			zone = x[:len(z)]
			break
		}
	}
	if zone == "" && strings.HasPrefix(x, "04") {
		zone = x[:4]
	}
	x = x[len(zone):]
	nobreak := "\u00A0"
	switch {
	case zone == "0903":
		x = x[:2] + nobreak + x[2:]
	case len(x) == 6:
		x = x[:2] + nobreak + x[2:4] + nobreak + x[4:]
	default:
		x = x[:3] + nobreak + x[3:5] + nobreak + x[5:]
	}
	return zone + nobreak + x
}

// Phone transformation
func Phone(body string) string {

	// 09 385 62 03
	// 0475 812 419
	rex := regexp.MustCompile(`[0-9][0-9./ -]{8,}`)

	phones := rex.FindAllString(body, -1)
	if len(phones) == 0 {
		return body
	}
	result := body
	rex2 := regexp.MustCompile(`[^0-9]`)
	rex3 := regexp.MustCompile(`[^0-9]+$`)
	for _, phone := range phones {
		if !strings.HasPrefix(phone, "0") {
			continue
		}
		phone = rex3.ReplaceAllString(phone, "")
		x := rex2.ReplaceAllString(phone, "")
		if len(x) == 9 || len(x) == 10 {
			result = strings.ReplaceAll(result, phone, showphone(x))
		}
	}
	return result
}

func Nobreak(s string) string {
	return strings.ReplaceAll(s, " ", "\u00A0")
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