package tools

import (
	"bytes"
	"strings"
	"unicode"

	"brocade.be/pbladng/lib/registry"
	unidecode "github.com/mozillazg/go-unidecode"

	"golang.org/x/text/unicode/norm"
)

var chars = map[rune]string{}

func init() {
	chrs := registry.Registry["characters"].(map[string]any)
	for r, v := range chrs {
		c := []rune(r)
		if len(c) != 0 {
			chars[c[0]] = v.(string)
		}
	}
}
func Latin1(s string) string {
	s = norm.NFC.String(s)
	latin1 := true
	for _, c := range s {
		if c > unicode.MaxLatin1 {
			latin1 = false
		}
	}
	if latin1 {
		return s
	}
	var buffer bytes.Buffer
	var r rune = 'A'
	for i, c := range s {
		if c <= unicode.MaxLatin1 {
			if r == ' ' && c == ' ' {
				continue
			}
			buffer.WriteRune(c)
			if c == ' ' {
				r = ' '
			} else {
				r = 'A'
			}
			continue
		}
		ch, ok := chars[c]
		if !ok {
			ch = unidecode.Unidecode(string(c))
		}

		if strings.HasPrefix(ch, " ") && (i == 0 || (r == ' ' && ch != " ")) {
			ch = ch[1:]
		}
		buffer.WriteString(ch)
		r = 'A'
		if strings.HasSuffix(ch, " ") {
			r = ' '
		}

	}
	return buffer.String()
}
