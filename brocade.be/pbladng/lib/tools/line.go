package tools

import (
	"fmt"
	"strings"
)

func FixColon(s string) string {
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
		k := strings.IndexAny(part, "1234567890")
		if k == 0 {
			continue
		}
		parts[i] = " " + strings.TrimLeft(parts[i], " \t")
		ppart = parts[i]
	}
	return strings.Join(parts, ":")
}

func FixDelim(s string, index int, sub string, r rune) string {
	if strings.HasPrefix(strings.TrimSpace(s), "//") {
		return s
	}
	prefix := ""
	if index > 0 {
		prefix = s[:index]
	}
	sr := string(r)
	suffix := s[index+len(sub):]
	if Count(prefix, r)%2 == 0 {
		return prefix + sr + sub + sr + suffix
	}
	if Count(suffix, r)%2 == 0 {
		return prefix + sub + sr + suffix
	}
	return prefix + sub + suffix
}

func GetDelims(s string, r rune, m map[string]bool) {
	if !strings.ContainsRune(s, r) {
		return
	}
	parts := strings.SplitN(s, string(r), -1)
	pieces := make([]string, 0)
	for i, part := range parts {
		if i == 0 {
			pieces = append(pieces, part)
			continue
		}
		piece := pieces[len(pieces)-1]
		k := len(piece)
		l := len(strings.TrimRight(piece, "\\"))
		if (k-l)%2 == 1 {
			pieces[len(pieces)-1] = piece + string(r) + part
			continue
		}
		pieces = append(pieces, part)
	}

	for i, piece := range pieces {
		if i%2 == 0 {
			continue
		}
		piece = strings.TrimSpace(piece)
		if piece != "" {
			m[piece] = true
		}
	}
}

func Count(s string, r rune) int {
	if !strings.ContainsRune(s, r) {
		return 0
	}
	sr := string(r)
	s = strings.ReplaceAll(s, `\\`, "")
	s = strings.ReplaceAll(s, `\`+sr, "")
	return strings.Count(s, sr)
}

func IndexRune(s string, r rune) (pos int) {
	k := strings.IndexRune(s, r)
	if k == -1 {
		return -1
	}
	prefix := s[:k]
	l := len(strings.TrimRight(prefix, "\\"))
	if (k-l)%2 == 0 {
		return k
	}
	s = s[k+1:]
	l = IndexRune(s, r)
	if l == -1 {
		return -1
	}
	return k + l + 1
}

func Protected(s string) []string {
	k := IndexRune(s, '{')
	if k == -1 {
		return []string{s}
	}
	l := IndexRune(s[k+1:], '}')
	if l == -1 {
		return []string{s}
	}
	fmt.Println("s:", s[k+1:], l)
	return append([]string{s[:k], s[k : k+l+2]}, Protected(s[k+l+2:])...)
}
