package tools

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	bstrings "brocade.be/base/strings"
)

var phonerex = regexp.MustCompile(`[0-9][0-9./ -]{8,}`)
var nondigit = regexp.MustCompile(`[^0-9]`)
var endnondigit = regexp.MustCompile(`[^0-9]+$`)
var spaces = regexp.MustCompile("  +")
var ws = regexp.MustCompile(`\s`)

func Heading(s string) string {
	k := strings.Index(s, "[")
	if k != -1 {
		s = strings.TrimSpace(s[:k])
	}
	s = NormalizeH(s)
	return strings.TrimSpace(strings.ToUpper(s))
}

func NormalizeH(s string) string {

	s = DeleteMeta(s)
	s = bstrings.InsertDiacritic(s)
	s = Euro(s)
	s = Phone(bstrings.Latin1(s))
	s = Colon(s)
	s = spaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func Normalize(s string, trim bool) (string, string) {
	if strings.HasPrefix(s, "=") {
		return bstrings.Latin1(s), ""
	}
	s = bstrings.InsertDiacritic(s)
	s = Euro(s)
	s = Phone(bstrings.Latin1(s))
	s = Bank(s, true)
	s = Hour(s)
	s, maxday := Day(s, true)
	s = Colon(s)
	s = spaces.ReplaceAllString(s, " ")
	if !trim {
		return s, maxday
	}
	return strings.TrimSpace(s), maxday
}

func NormalizeR(s string, trim bool) string {
	if strings.TrimSpace(s) == "-" {
		return "-"
	}
	if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "--") && !strings.HasPrefix(s, "- ") {
		s = "- " + s[1:]
	}

	if strings.HasPrefix(s, "=") {
		return s
	}
	if strings.HasPrefix(s, "-") {
		return s
	}
	s = bstrings.InsertDiacritic(s)
	s = Euro(s)
	s = Phone(bstrings.Latin1(s))
	if !trim {
		return s
	}
	return strings.TrimSpace(s)
}

func DeleteMeta(s string) string {
	set := "*_|"
	if !strings.ContainsAny(s, set) {
		return s
	}
	s = strings.ReplaceAll(s, "\\\\", "\x00")
	for _, r := range set {
		if !strings.ContainsRune(s, r) {
			continue
		}
		rs := `\` + string(r)
		s = strings.ReplaceAll(s, rs, "\x03")
		s = strings.ReplaceAll(s, string(r), "")
		s = strings.ReplaceAll(s, "\x03", rs)
	}
	return strings.ReplaceAll(s, "\x00", "\\\\")
}

func MetaChars(s string) string {

	set := "*#`"
	found := ""
	s = strings.ReplaceAll(s, "\\\\", "")
	for _, r := range set {
		if !strings.ContainsRune(s, r) {
			continue
		}
		rs := `\` + string(r)
		s = strings.ReplaceAll(s, rs, "")
		if !strings.ContainsRune(s, r) {
			continue
		}
		found += string(r)
	}
	return found
}

func Phone(s string) string {

	// 09 385 62 03
	// 0475 812 419
	phones := phonerex.FindAllString(s, -1)
	if len(phones) == 0 {
		return s
	}
	result := s
	for _, phone := range phones {
		if !strings.HasPrefix(phone, "0") {
			continue
		}
		phone = endnondigit.ReplaceAllString(phone, "")
		x := nondigit.ReplaceAllString(phone, "")
		if len(x) == 9 || len(x) == 10 {
			result = strings.ReplaceAll(result, phone, showphone(x))
		}
	}
	return result
}

func Colon(s string) string {
	if strings.HasPrefix(strings.TrimSpace(s), "//") {
		return s
	}
	if !strings.ContainsRune(s, ':') {
		return s
	}
	parts := strings.SplitN(s, ":", -1)
	ppart := ""
	for i, part := range parts {
		if i == 0 {
			ppart = part
			continue
		}
		if strings.HasSuffix(ppart, "http") || strings.HasSuffix(ppart, "https") || strings.HasSuffix(ppart, "mailto") {
			ppart = part
			continue
		}
		ppart = part
		k := strings.IndexAny(part, "1234567890`*")
		if k == 0 {
			continue
		}
		parts[i] = " " + strings.TrimLeft(parts[i], " \t")
		ppart = parts[i]
	}
	return strings.Join(parts, ":")
}

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
	x = nondigit.ReplaceAllString(x, "")
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

func Isdash(s string) bool {
	s = strings.ReplaceAll(s, "_", "-")
	s = ws.ReplaceAllString(s, "")
	if len(s) < 5 {
		return false
	}
	s = strings.TrimLeft(s, "-")
	return s == ""
}

func WSLines(mylist []string) (result []string) {
	found := false
	prev := false
	last := -1
	for _, line := range mylist {
		s := strings.TrimSpace(line)
		if !found && s == "" {
			continue
		}
		found = true
		if prev && s == "" {
			continue
		}
		result = append(result, s)
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

func YesNo(s string) bool {
	for {
		fmt.Printf("%s [y/n] ", s)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		if strings.HasPrefix(text, "y") {
			return true
		}
		if strings.HasPrefix(text, "n") {
			return false
		}
	}
}

func IsTrue(s string) bool {
	if s == "" {
		return false
	}
	s = strings.TrimSpace(strings.ToLower(s))
	return !strings.HasPrefix(s, "n")
}

func CheckBalancedChar(s string, r rune) bool {
	if !strings.ContainsRune(s, r) {
		return true
	}
	if strings.ContainsRune(s, '\\') {
		s = strings.ReplaceAll(s, "\\\\", "")
		s = strings.ReplaceAll(s, string('\\')+string(r), "")
	}
	count := strings.Count(s, string(r))
	return count%2 == 0
}

func CheckBalanced(s string) string {
	if !strings.ContainsAny(s, "|*_") {
		return ""
	}
	if strings.ContainsRune(s, '\\') {
		s = strings.ReplaceAll(s, "\\\\", "")
		for _, ch := range "|*_" {
			s = strings.ReplaceAll(s, string('\\')+string(ch), "")
		}
	}
	again := make([]rune, 0)
	for _, ch := range "|*_" {
		count := strings.Count(s, string(ch))
		if count%2 != 0 {
			return string(ch)
		}
		if count != 0 {
			again = append(again, ch)
		}
	}
	if len(again) < 2 {
		return ""
	}
	for _, r1 := range again {
		for _, r2 := range again {
			if r1 == r2 {
				continue
			}
			parts := strings.SplitN(s, string(r1), -1)
			parts = parts[1 : len(parts)-1]
			for _, part := range parts {
				if !CheckBalancedChar(part, r2) {
					return string(r1) + string(r2) + " :" + part
				}
			}
		}
	}
	return ""
}

func DoubleChar(s string, r rune) string {
	sub := string(r) + string(r)
	if !strings.Contains(s, sub) {
		return s
	}
	ok := false
	if strings.ContainsRune(s, '\\') {
		ok = true
		s = strings.ReplaceAll(s, "\\\\", string([]byte{0, 0}))
		s = strings.ReplaceAll(s, string('\\')+string(r), string([]byte{0, 3}))
	}
	s = strings.ReplaceAll(s, sub, "")
	if !ok {
		return s
	}
	return strings.ReplaceAll(strings.ReplaceAll(s, string([]byte{0, 3}), string('\\')+string(r)), string([]byte{0, 0}), "\\\\")

}
