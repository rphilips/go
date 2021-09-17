package yottadb

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	"lang.yottadb.com/go/yottadb"
)

var Space string = strings.Repeat(strings.Repeat(strings.Repeat(" ", 64), 32), 32)
var Lit1 = regexp.MustCompile(`^"[^"]*$"`)
var Number3 = regexp.MustCompile(`^[+-]?(([0-9]+(\.[0-9]+)?)|(\.[0-9]+))(E[+-]?[0-9]+)$`)
var Number4 = regexp.MustCompile(`^[+-]?[0-9]+(\.[0-9]+)?$`)
var Number1 = regexp.MustCompile(`^[+-]?[0-9]+$`)
var Number2 = regexp.MustCompile(`^[+-]?[0-9]+E[+]?[0-9]+$`)

func QS(glvn string) (subs []string) {

	glvn = strings.TrimSpace(glvn)
	if glvn == "" {
		return nil
	}
	k := strings.Index(glvn, "(")
	if k == -1 {
		subs = append(subs, glvn)
		return
	}
	subs = append(subs, glvn[:k])
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
			even = !even
			sub += string(r)
			continue
		case '(':
			if even {
				level++
			}
			sub += string(r)
			continue
		case ')':
			if even {
				level--
			}
			sub += string(r)
			continue
		case ' ':
			if !even {
				sub += string(r)
			}
			continue
		case ',':
			if !even || level != 0 {
				sub += string(r)
				continue
			}
			subs = append(subs, sub)
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
		if level != 0 {
			sub += strings.Repeat(")", level)
		}
		subs = append(subs, sub)
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
		args := make([]string, len(subs)-1)
		for i, sub := range subs {
			if i != 0 {
				if !isliteral(sub) {
					exec := `s %qzxy0=` + sub
					r, err := Calc(exec, `%qzxy0`)
					if err == nil {
						sub = r
						if !isNumber(sub) {
							sub = `"` + strings.ReplaceAll(sub, `"`, `""`) + `"`
						}
					} else {
						sub = err.Error()
					}
				}
				args[i-1] = sub
			}
		}
		return subs[0] + `(` + strings.Join(args, ",") + `)`
	}
}

func Glvn(ref string) (glvn string) {
	oref := ref
	if ref == "" {
		return oref
	}
	global := strings.HasPrefix(ref, "/") || strings.HasPrefix(ref, "^")
	if global {
		ref = ref[1:]
	}
	if ref == "" {
		return oref
	}
	rex := regexp.MustCompile(`^[%A-Za-z][A-Za-z0-9]*`)
	k := strings.IndexAny(ref, "/(")
	if k == -1 {
		if !rex.MatchString(ref) {
			return oref
		}
		if global {
			return "^" + ref
		}
		return ref
	}
	notation := rune(ref[k])

	name := ref[:k]
	ref = ref[k+1:]
	if global {
		name = "^" + name
	}
	subs := []string{name}
	switch notation {
	case '(':
		glvn := name + `(` + ref
		subs = QS(glvn)
	case '/':
		ref = strings.TrimSpace(ref)
		if ref == "" {
			return oref
		}
		ref = qutil.Escape(ref)
		parts := strings.SplitN(ref, "/", -1)
		for _, part := range parts {
			part = qutil.Unescape(part)
			subs = append(subs, part)
		}
	}
	nemps := make([]string, 0)
	for _, sub := range subs {
		if sub != "" {
			nemps = append(nemps, sub)
		}
	}

	if len(nemps) < 2 {
		return UnQS(subs)
	}
	mvar := UnQS(nemps)
	exec := `s %qzxy0=$NA(` + mvar + `)`
	r, err := Calc(exec, `%qzxy0`)
	if r == "" || err != nil {
		return UnQS(subs)
	}
	rsubs := QS(r)
	j := 0
	for i, sub := range subs {
		if sub != "" {
			subs[i] = rsubs[j]
			j++
		}
	}
	return UnQS(subs)
}

// D return $D() of reference, (value of -1 for iinvalid reference)
func D(glvn string) int {
	subs := QS(glvn)
	if len(subs) == 0 {
		return -1
	}
	d, err := yottadb.DataE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if err != nil {
		return -1
	}
	return int(d)
}

// G return $G() of reference, but returns an error
func G(glvn string) (value string, err error) {
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", errors.New("invaid reference")
	}
	value, err = yottadb.ValE(yottadb.NOTTP, nil, subs[0], subs[1:])
	return
}

// Set
func Set(glvn string, value string) (err error) {
	subs := QS(glvn)
	if len(subs) == 0 {
		return errors.New("invaid reference")
	}
	return yottadb.SetValE(yottadb.NOTTP, nil, value, subs[0], subs[1:])
}

// SetArg decompses an argument for the set command 's glvn=arg'
func SetArg(arg string) (glvn string, value string, err error) {
	if !strings.ContainsRune(arg, '=') {
		arg = Glvn(arg)
		value, err = G(arg)
		if err != nil {
			return arg, "", errors.New("contains no '='")
		}
		return arg, value, nil
	}
	var parts []string
	k := strings.IndexAny(arg, "/(=")
	r := rune(arg[k])
	var e error
	switch r {
	case '=':
		parts = strings.SplitN(arg, "=", 2)
		glvn = Glvn(parts[0])
		value = parts[1]
		return glvn, value, nil
	case '/':
		parts, e = splits(arg)
		if e != nil {
			return arg, "", e
		}
		glvn = Glvn(parts[0])
		value = parts[1]
		return glvn, value, nil
	default:
		parts, e = splitb(arg)
		if e != nil {
			return arg, "", e
		}
		glvn = Glvn(parts[0])
		value = parts[1]
		return glvn, value, nil
	}

}

// Next = $Next
func Next(glvn string) (nglvn, next string, err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", "", errors.New("cannot determine next")
	}
	z, e := yottadb.SubNextE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
		return "", "", errors.New("cannot determine next")
	}
	subs[len(subs)-1] = z
	return UnQS(subs), z, nil
}

// Prev = $Next(-1)
func Prev(glvn string) (pglvn string, value string, err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", "", errors.New("cannot determine prev")
	}
	z, e := yottadb.SubPrevE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
		return "", "", errors.New("cannot determine prev")
	}
	subs[len(subs)-1] = z
	return UnQS(subs), z, nil
}

func Right(glvn string) (rglvn string, err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", errors.New("cannot determine right")
	}

	if len(subs) > 1 && subs[len(subs)-1] == "" {
		return glvn, err
	}
	subs = append(subs, "")
	return UnQS(subs), nil
}

func Left(glvn string) (rglvn string, err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", errors.New("cannot determine left")
	}
	if len(subs) == 1 {
		return glvn, nil
	}
	subs = subs[:len(subs)-1]
	return UnQS(subs), nil
}

func KillN(glvn string) (err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return errors.New("cannot determine node")
	}
	err = yottadb.DeleteE(yottadb.NOTTP, nil, yottadb.YDB_DEL_NODE, subs[0], subs[1:])
	return err
}

func KillT(glvn string) (err error) {
	glvn = Glvn(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return errors.New("cannot determine tree")
	}
	err = yottadb.DeleteE(yottadb.NOTTP, nil, yottadb.YDB_DEL_TREE, subs[0], subs[1:])
	return err
}

func splits(glvn string) (parts []string, e error) {
	glvn = qutil.Escape(glvn)
	if !strings.ContainsRune(glvn, '=') {
		value, err := G(glvn)
		if err != nil {
			return nil, errors.New("contains no '='")
		}
		glvn += "=" + qutil.Escape(value)
	}
	parts = strings.SplitN(glvn, "=", 2)
	parts[0] = qutil.Unescape(parts[0])
	parts[1] = qutil.Unescape(parts[1])
	return parts, nil
}

func splitb(glvn string) (parts []string, e error) {
	if strings.HasPrefix(glvn, "=") {
		return nil, errors.New("contains no reference part")
	}
	offset := 0
	for {
		k := strings.IndexByte(glvn[offset:], '=')
		if k == -1 {
			g := Glvn(glvn)
			value, err := G(glvn)
			if err != nil {
				return nil, errors.New("contains no '='")
			}
			parts[0] = g
			parts[1] = value
			return parts, nil
		}
		before := glvn[:offset+k]
		if strings.Count(before, `"`)%2 == 1 {
			offset += k + 1
			continue
		}
		parts[0] = before
		parts[1] = glvn[k+1:]
		return parts, nil
	}
}

func Exit() {
	yottadb.Exit()
}

func Exec(text string) error {
	fmtable := "/home/rphilips/.yottadb/ydbaccess.ci"
	envvarSave := make(map[string]string)
	qutil.SaveEnvvars(&envvarSave, "ydb_ci", "ydb_routines")
	os.Setenv("ydb_ci", fmtable)
	out := Space
	_, err := yottadb.CallMT(yottadb.NOTTP, nil, 0, "xecute", text, &out)
	qutil.RestoreEnvvars(&envvarSave, "ydb_ci", "ydb_routines")
	if err != nil {
		return fmt.Errorf("Exec error: %s", err.Error())
	}
	return nil
}

func Calc(text string, glvn string) (string, error) {
	err := Exec(text)
	if err != nil {
		return "", err
	}
	return G(glvn)
}

func isliteral(term string) bool {
	if term == `""` {
		return true
	}
	if strings.HasPrefix(term, `"`) {
		if Lit1.MatchString(term) {
			return true
		}
		if strings.HasSuffix(term, `"`) {
			return false
		}
		term = strings.TrimPrefix(term, `"`)
		term = strings.TrimSuffix(term, `"`)
		term = `"` + strings.ReplaceAll(term, `""`, ``) + `"`
		return Lit1.MatchString(term)
	}
	if Number1.MatchString(term) {
		return true
	}
	return false
}

func isNumber(term string) bool {
	if Number1.MatchString(term) {
		return true
	}
	if Number4.MatchString(term) {
		return true
	}
	return false
}
