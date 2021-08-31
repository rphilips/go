package mumps

import (
	"errors"
	"regexp"
	"strings"

	"lang.yottadb.com/go/yottadb"
)

func GlobalRef(glo string) (gloref string, subs []string, err error) {

	glo = strings.TrimSpace(glo)
	k := strings.IndexAny(glo, "/(")
	if k != -1 && glo[k:k+1] == "(" && !strings.HasSuffix(glo, ")") {
		glo += ")"
	}
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
		return "", nil, errors.New("reference contains \\x01 character")
	}

	rex := regexp.MustCompile(`^[%A-Za-z][A-Za-z0-9]*\(`)
	if rex.MatchString(glo) {
		glo = splitGlo(glo)
	}
	glo = Escape(glo)
	subs = strings.SplitN(glo, "/", -1)
	name := subs[0]
	rex = regexp.MustCompile("^[%A-Za-z][A-Za-z0-9]*$")
	if !rex.MatchString(name) {
		return "", nil, errors.New("invalid reference")
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
	for i, sub := range subs {
		if i == 0 {
			continue
		}
		if i != 1 {
			gloref += ","
		}
		sub, n, _ := Nature(sub)
		sub = Unescape(sub)
		subs[i] = sub
		if n == "s" {
			sub = strings.ReplaceAll(sub, `"`, `""`)
			gloref += `"` + sub + `"`
		} else {
			gloref += subs[i]
		}
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

func EditGlobal(glo string) (gloref string, value string, err error) {
	if !strings.ContainsRune(glo, '=') {
		value, err = getglo(glo)
		if err != nil {
			return glo, "", errors.New("contains no '='")
		}
		glo += "=" + value
	}
	k := strings.IndexAny(glo, "/(")
	var parts []string
	if k == -1 {
		parts = strings.SplitN(glo, "=", 2)
	} else {
		var e error
		if glo[k:k+1] == "/" {
			parts, e = splits(glo)
		} else {
			parts, e = splitb(glo)
		}
		if e != nil {
			return "", "", e
		}
	}
	gloref = strings.TrimSpace(parts[0])
	gloref, _, err = GlobalRef(gloref)
	if err != nil {
		return "", "", err
	}
	value, _, _ = Nature(parts[1])
	return gloref, value, nil
}

func splits(glo string) (parts []string, e error) {
	glo = Escape(glo)
	if !strings.ContainsRune(glo, '=') {
		value, err := getglo(glo)
		if err != nil {
			return nil, errors.New("contains no '='")
		}
		glo += "=" + Escape(value)
	}
	parts = strings.SplitN(glo, "=", 2)
	parts[0] = Unescape(parts[0])
	parts[1] = Unescape(parts[1])
	return parts, nil
}

func splitb(glo string) (parts []string, e error) {
	glo = Escape(glo)
	if strings.HasPrefix(glo, "=") {
		return nil, errors.New("contains no reference part")
	}
	offset := 0
	for {
		k := strings.IndexByte(glo[offset:], '=')
		if k == -1 {
			value, err := getglo(glo)
			if err != nil {
				return nil, errors.New("contains no '='")
			}
			glo += "=" + Escape(value)
			continue
		}
		if k == 0 {
			return nil, errors.New("contains no reference part")
		}
		before := glo[:offset+k]
		if strings.Count(before, `"`)%2 == 0 {
			parts = append(parts, Unescape(before), Unescape(glo[k+offset+1:]))
			return parts, nil
		}
		offset = offset + k + 1
	}
}

func getglo(glo string) (value string, err error) {
	_, subs, err := GlobalRef(glo)
	if err != nil {
		return
	}
	value, err = yottadb.ValE(yottadb.NOTTP, nil, subs[0], subs[1:])
	return
}
