package util

import (
	"strings"
)

func MName(glvn string, glo bool) string {
	subs := QS(glvn)
	name := UnQS(subs)
	if glo && name != "" && len(subs) != 0 && !strings.Contains(subs[0], ":") {
		name := strings.TrimLeft(name, "^")
		if name != "" {
			name = "^" + name
			return name
		}
	}
	return name
}

func QS(glvn string) (subs []string) {

	glvn = strings.TrimSpace(glvn)
	if glvn == "" {
		return nil
	}

	k := strings.IndexAny(glvn, "(/")
	if k == -1 {
		subs = append(subs, glvn)
		return
	}

	subs = append(subs, glvn[:k])

	if glvn[k:k+1] == "/" {
		parts := strings.SplitN(glvn[k+1:], "/", -1)
		if len(parts) == 0 {
			return
		}
		for _, part := range parts {
			subs = append(subs, `"`+strings.ReplaceAll(part, `"`, `""`)+`"`)
		}
		return
	}
	glvn = strings.TrimSuffix(glvn[k+1:], ")")
	glvn = strings.TrimSpace(glvn)
	glvn = strings.TrimSuffix(glvn, ",")

	if glvn == "" {
		return nil
	}
	level := 0
	sub := ""
	even := true
	for _, r := range glvn {
		if r < 32 || r > 127 {
			sub += string(r)
			continue
		}
		switch r {
		case '"':
			sub += string(r)
			even = !even
			continue
		case '(':
			sub += string(r)
			if even {
				level++
			}
			continue
		case ')':
			sub += string(r)
			if even {
				level--
			}
			continue
		case ',':
			if !even || level != 0 {
				sub += string(r)
				continue
			}
			subs = append(subs, sub)
			level = 0
			even = true
			sub = ""
			continue
		default:
			sub += string(r)
			continue
		}
	}

	if sub != "" {
		if strings.Count(sub, `"`)%2 == 1 {
			sub += `"`
		}
		if level > 0 {
			sub += strings.Repeat(")", level)
		}
		subs = append(subs, sub)
	}
	for i, sub := range subs {
		sub = strings.TrimSpace(sub)
		if !strings.HasPrefix(sub, `"`) && strings.Contains(sub, "(") {
			subsubs := QS(sub)
			sub = UnQS(subsubs)
		}
		subs[i] = sub
	}
	return
}

func UnQS(subs []string) (glvn string) {
	switch len(subs) {
	case 0:
		return ""
	case 1:
		return subs[0]
	default:
		return subs[0] + `(` + strings.Join(subs[1:], ",") + `)`
	}
}
