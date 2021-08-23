package mumps

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func GlobalRef(glo string) (gloref string, subs []string, err error) {

	glo = strings.TrimSpace(glo)
	global := strings.HasPrefix(glo, "/") || strings.HasPrefix(glo, "^")
	if global {
		glo = strings.TrimPrefix(glo, "/")
		glo = strings.TrimPrefix(glo, "^")
		glo = strings.TrimSpace(glo)
	}
	if glo == "" {
		return "", nil, errors.New("invalid global reference")
	}

	if strings.ContainsRune(glo, '\x00') {
		return "", nil, errors.New("global reference contains NUL character")
	}

	if strings.ContainsRune(glo, '\x01') {
		return "", nil, errors.New("global reference contains \\x01 character")
	}

	rex := regexp.MustCompile(`^[%A-Za-z][A-Za-z0-9]*\(`)
	if rex.MatchString(glo) {
		glo = splitGlo(glo)
	}
	glo = strings.ReplaceAll(glo, "\\\\", "\x00")
	glo = strings.ReplaceAll(glo, "\\/", "\x01")
	subs = strings.SplitN(glo, "/", -1)
	name := subs[0]
	rex = regexp.MustCompile("^[%A-Za-z][A-Za-z0-9]*$")
	if !rex.MatchString(name) {
		return "", nil, errors.New("invalid global reference")
	}
	gloref = ""
	if global {
		gloref = "^"
	}
	gloref += name
	subs[0] = gloref
	if len(subs) == 1 {
		return
	}
	gloref += "("
	number1 := regexp.MustCompile(`^[+-]?(([0-9]+(\.[0-9]+)?)|(\.[0-9]+))(E[+-]?[0-9]+)$`)
	number2 := regexp.MustCompile(`^[+-]?[0-9]+$`)
	number3 := regexp.MustCompile(`^[+-]?[0-9]+E[+]?[0-9]+$`)
	for i, sub := range subs {
		if i == 0 {
			continue
		}
		if i != 1 {
			gloref += ","
		}
		switch {

		case number3.MatchString(sub):
			x := strings.SplitN(sub, "E", -1)
			y := strings.ReplaceAll(x[len(x)-1], "+", "")
			z, err := strconv.ParseInt(y, 10, 0)
			if err != nil || z > 64 {
				gloref += sub
				continue
			}
			y = x[0] + strings.Repeat("0", int(z))
			if _, err := strconv.ParseInt(y, 10, 64); err == nil {
				gloref += y
				subs[i] = y
				continue
			}
		case number2.MatchString(sub):
			if _, err := strconv.ParseInt(sub, 10, 64); err == nil {
				gloref += sub
				continue
			}
		case number1.MatchString(sub):
			if x, err := strconv.ParseFloat(sub, 64); err == nil {
				y := fmt.Sprintf("%f", x)
				gloref += y
				continue
			}
		default:
			sub = strings.ReplaceAll(sub, "\x00", `\`)
			sub = strings.ReplaceAll(sub, "\x01", `/`)
		}
		subs[i] = sub
		sub = strings.ReplaceAll(sub, `"`, `""`)
		gloref += `"` + sub + `"`
	}
	gloref += ")"
	return

}

func splitGlo(glo string) string {
	k := strings.Index(glo, "(")
	name := strings.TrimSpace(glo[:k])
	rest := strings.TrimSpace(glo[k+1:])
	rest = strings.TrimSuffix(rest, ")")

	parts := []string{name}
	before := ""
	for {
		k := strings.Index(rest, ",")
		if k < 0 {
			before += rest
			rest = ""
		} else {
			before += rest[:k]
			rest = rest[k+1:]
		}
		if rest != "" && strings.Count(before, "\"")%2 != 0 {
			before += ","
			continue
		}
		if rest == "" || strings.Count(before, "\"")%2 == 0 {
			sub := strings.TrimSpace(before)
			before = ""
			sub = strings.TrimPrefix(sub, `"`)
			sub = strings.TrimSuffix(sub, `"`)
			sub = strings.ReplaceAll(sub, `""`, `"`)
			sub = strings.ReplaceAll(sub, `\`, `\\`)
			sub = strings.ReplaceAll(sub, `/`, `\/`)
			parts = append(parts, sub)
		}
		if rest == "" {
			break
		}
	}
	return strings.Join(parts, "/")
}
