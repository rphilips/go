package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var rhour = regexp.MustCompile(`\b(([0-9]{1,2}\.[0-9]{1,2} ?[hu]\.?)|([0-9]{1,2}[uh][0-9]{1,2})|([0-9]{1,2} ?[uh]\.?))\b`)
var rday = regexp.MustCompile(`\b(([0-9]{1,2}[-./][0-9]{1,2}[-./][0-9]{2,4})?)|([0-9]{1,2}/[0-9]{1,2})|([0-9]{1,2} ?[a-zA-Z.]{3,10}( '?[0-9]{2,4})?))\b`)

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

func Day(s string, bolden bool) string {
	parts := rhour.FindAllStringIndex(s, -1)
	if parts == nil {
		return s
	}


	result := make([]string, 0)
	for i := 0; i < len(parts); i++ {
		day := s[parts[i][0]:parts[i][1]]
		after := ""
		if i+1 < len(parts) {
			after = s[parts[i][1]:parts[i+1][0]]
		} else {
			after = s[parts[i][1]:]
		}






		m := 0
		x := strings.TrimLeft(day, "1234567890")
		d, e := strconv.Atoi(day[:len(day)-len(x)])
		if e != nil || d > 31 {
			result = append(result, day)
			result = append(result, after)
			continue
		}
		rest := day[len(day)-len(x):]
		rest = strings.TrimLeft(rest, " /.-")
		x = strings.TrimLeft(rest, "1234567890")
		if x != rest {
			m, e = strconv.Atoi(rest[:len(rest)-len(x)])
			if e != nil || m > 12 {
				result = append(result, day)
				result = append(result, after)
				continue
			}
		} else {
			x = strings.TrimLeft(rest, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			if x == rest {
				result = append(result, day)
				result = append(result, after)
				continue
			}
			month := strings.ToLower(rest[:len(rest)-len(x)])
			now := time.now()
			z :=
		}

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
	return s
}
