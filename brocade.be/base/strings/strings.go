package strings

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func LeftTrimSpace(s string) string {
	x := strings.TrimSpace(s)
	if x == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(x)
	k := strings.IndexRune(s, r)
	return s[k:]
}

func RightTrimSpace(s string) string {
	x := strings.TrimSpace(s)
	if x == "" {
		return ""
	}
	r, size := utf8.DecodeLastRuneInString(x)
	k := strings.LastIndexAny(s, string(r))
	return s[:k+size]
}

func LeftRuneString(s string) string {
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return ""
	}
	return string(r)
}

func RightRuneString(s string) string {
	if s == "" {
		return ""
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	if r == utf8.RuneError {
		return ""
	}
	return string(r)
}

var replacers = map[string]*strings.Replacer{}

func Template(s string, keys map[string]string, prefix string, suffix string, id string) string {

	if s == "" || len(keys) == 0 {
		return ""
	}
	r := replacers[id]
	if r == nil {
		slice := make([]string, 2*len(keys))
		for k, v := range keys {
			slice = append(slice, prefix+k+suffix, v)
		}
		r = strings.NewReplacer(slice...)
		if id != "" {
			replacers[id] = r
		}
	}
	return r.Replace(s)
}

func JSON[J string | []string | map[string]string](data ...J) string {
	if len(data) == 0 {
		return ""
	}
	if len(data) == 1 {
		blob, _ := json.MarshalIndent(data[0], "", "    ")
		return string(blob)
	}

	blob, _ := json.MarshalIndent(data, "", "    ")
	return string(blob)
}
