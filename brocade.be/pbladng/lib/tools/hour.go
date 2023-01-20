package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	btime "brocade.be/base/time"
)

var rhour = regexp.MustCompile(`\b(([0-9]{1,2}\.[0-9]{1,2} ?[hu]\.?)|([0-9]{1,2}[uh][0-9]{1,2})|([0-9]{1,2} ?[uh]\.?))\b`)
var rday = regexp.MustCompile(`\b(([0-9]{1,2}[-./][0-9]{1,2}[-./][0-9]{2,4})|([0-9]{1,2}/[0-9]{1,2})|([0-9]{1,2} ?[a-zA-Z.]{3,10}( '?[0-9]{2,4})?))\b`)
var ryear = regexp.MustCompile(`([0-9]{4})|('[0-9]{2})`)

func Hour(s string) string {

	parts := rhour.FindAllStringIndex(s, -1)
	if parts == nil {
		return s
	}
	result := make([]string, 0)
	if parts[0][0] != 0 {
		result = append(result, s[0:parts[0][0]])
	}
	for i := 0; i < len(parts); i++ {
		hour := s[parts[i][0]:parts[i][1]]
		after := ""
		if i+1 < len(parts) {
			after = s[parts[i][1]:parts[i+1][0]]
		} else {
			after = s[parts[i][1]:]
		}
		m := 0
		x := strings.TrimLeft(hour, "1234567890")
		h, e := strconv.Atoi(hour[:len(hour)-len(x)])
		if e != nil || h > 24 {
			result = append(result, hour)
			result = append(result, after)
			continue
		}
		rest := strings.ReplaceAll(hour[len(hour)-len(x):], "h", "u")
		if strings.TrimLeft(rest, " u.") == "" {
			result = append(result, fmt.Sprintf("%02d.%02d u.\x00", h, m))
			result = append(result, after)
			continue
		}
		rest = rest[1:]
		x = strings.TrimLeft(rest, "1234567890")
		m, e = strconv.Atoi(rest[:len(rest)-len(x)])
		if e != nil || m > 59 {
			result = append(result, hour)
			result = append(result, after)
			continue
		}
		result = append(result, fmt.Sprintf("%02d.%02d u.\x00", h, m))
		result = append(result, after)
		continue
	}
	r := strings.Join(result, "")
	r = strings.TrimRight(r, "\x00")
	r = strings.ReplaceAll(r, "\x00 .", ".")
	r = strings.ReplaceAll(r, "\x00 ", " ")
	r = strings.ReplaceAll(r, "\x00", " ")
	r = strings.ReplaceAll(r, "...", "\x00")
	r = strings.ReplaceAll(r, "..", ".")
	r = strings.ReplaceAll(r, ". .", ".")
	r = strings.ReplaceAll(r, "\x00", "...")
	return r
}

func Day(s string, bolden bool) (string, string) {
	parts := rday.FindAllStringIndex(s, -1)
	if parts == nil {
		return s, ""
	}
	maxday := ""
	result := ""
	if parts[0][0] != 0 {
		result += s[0:parts[0][0]]
	}
	for i := 0; i < len(parts); i++ {
		day := s[parts[i][0]:parts[i][1]]
		after := ""
		if i+1 < len(parts) {
			after = s[parts[i][1]:parts[i+1][0]]
		} else {
			after = s[parts[i][1]:]
		}
		tim := btime.DetectDate(day)
		if tim == nil {
			result += day + after
			continue
		}
		z := btime.StringDate(tim, "I")
		if z > maxday {
			maxday = z
		}

		mode := "X"
		if !ryear.MatchString(day) {
			mode = "NY"
		}
		sdate := btime.StringDate(tim, mode)
		if !bolden {
			result += sdate + "\x00" + after
			continue
		}

		x := strings.ReplaceAll(result, `\\`, "\x02")
		x = strings.ReplaceAll(x, `\*`, "\x04")
		if strings.Count(x, "*")%2 == 0 {
			result += "*" + sdate + "*\x00" + after
		} else {
			result += sdate + "\x00" + after
		}
	}
	result = strings.TrimRight(result, "\x00")
	result = strings.ReplaceAll(result, "\x00 .", ".")
	result = strings.ReplaceAll(result, "\x00 ", " ")
	result = strings.ReplaceAll(result, "\x00", "")
	result = strings.ReplaceAll(result, "...", "\x00")
	result = strings.ReplaceAll(result, "..", ".")
	result = strings.ReplaceAll(result, ". .", ".")
	result = strings.ReplaceAll(result, "\x00", "...")
	return result, maxday
}
