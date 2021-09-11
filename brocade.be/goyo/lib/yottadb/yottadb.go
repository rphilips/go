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

func QS(glvn string) (subs []string) {

	glvn = strings.TrimSpace(glvn)
	k := strings.Index(glvn, "(")
	if k == -1 {
		subs = append(subs, glvn)
		return
	}
	subs = append(subs, glvn[:k])
	glvn = strings.TrimSuffix(glvn[k+1:],")")
	glvn = strings.TrimSpace(glvn)
	glvn = strings.TrimSuffix(glvn,",")
	
	if glvn == "" {
		return
	}
	part := ""
	for {
		k := strings.Index(glvn, ",")
		if k == -1 {
			part += glvn
			glvn = ""
			part = strings.TrimSpace(part)
			subs = append(subs, part)
			part = ""
			break
		} else {
			part += ref[:k]
			glvn = ref[k+1:]
			if strings.Count(part, `"`)%2 == 1 {
				part += ","
				continue
			}
			part = strings.TrimSpace(part)
			subs = append(subs, part)
			part = ""
		}
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
				args[i-1] = `"` + strings.ReplaceAll(sub, `"`, `""`) + `"`
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
		ref = qutil.Escape(part)
		parts := strings.SplitN(ref, "/", -1)
		for _, part := range parts {
			part = strings.ReplaceAll(part, `"`, `""`)
			part = qutil.Unescape(part)
			subs = append(subs, `"`+part+`"`)
		}
	}
	nemps := make([]string, 0)
	for i, sub := range subs {
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
	if err == nil && r != "" {
		if last == `""` {
			r = strings.TrimSuffix(r, ")") + `,"")`
		}
		return r, nil
	}
	subs = append(subs, `""`)
	return subs[0] + "(" + strings.Join(subs[1:], ",") + ")", nil
}

// SplitGlo reduces global reference to /-form
func SplitGlvn(glo string) string {
	k := strings.Index(glo, "(")
	if k == -1 {
		return glo
	}

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











	
	glvnbal := strings.HasPrefix(glvn, "/") || strings.HasPrefix(glvn, "^")

	if glvnbal {
		glvn = strings.TrimPrefix(glvn, "/")
		glvn = strings.TrimPrefix(glvn, "^")
		glvn = strings.TrimSpace(glvn)
	}
	if glvn == "" {
		return "", nil, errors.New("invalid glvn reference")
	}

	if strings.ContainsRune(glvn, '\x00') {
		return "", nil, errors.New("glvn reference contains NUL character")
	}

	if strings.ContainsRune(glvn, '\x01') {
		return "", nil, errors.New("reference contains \\x01 character")
	}

	rex := regexp.MustCompile(`^[%A-Za-z][A-Za-z0-9]*\(`)
	if rex.MatchString(glvn) {
		glvn = SplitGlvn(glvn)
	}
	glvn = qutil.Escape(glvn)
	subs = strings.SplitN(glvn, "/", -1)
	name := subs[0]
	rex = regexp.MustCompile("^[%A-Za-z][A-Za-z0-9]*$")
	if !rex.MatchString(name) {
		return "", nil, errors.New("invalid reference")
	}
	ref = ""
	if glvnbal {
		ref = "^"
	}
	ref += name
	subs[0] = ref
	if len(subs) == 1 {
		return
	}
	ref += "("
	for i, sub := range subs {
		if i == 0 {
			continue
		}
		if i != 1 {
			ref += ","
		}
		sub, n, _ := qutil.Nature(sub)
		sub = qutil.Unescape(sub)
		subs[i] = sub
		if n == "s" {
			sub = strings.ReplaceAll(sub, `"`, `""`)
			ref += `"` + sub + `"`
		} else {
			ref += subs[i]
		}
	}
	ref += ")"
	return
}

func UnQS(subs []string) (glvn string) {
	switch len(subs) {
	case 0:
		return ""
	case 1:
		return subs[0]
	default:
		argums := make([]string, len(subs)-1)
		for i, sub := range subs {
			if i == 0 {
				continue
			}
			argums[i-1] = `"` + strings.ReplaceAll(sub, `"`, `""`) + `"`
		}
		return subs[0] + `(` + strings.Join(argums, `,`) + `)`
	}
}

// D return $D() of reference, (value of -1 for iinvalid reference)
func D(glvn string) int {
	_, subs, err := QS(glvn)
	if err != nil {
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
	_, subs, err := QS(glvn)
	if err != nil {
		return
	}
	value, err = yottadb.ValE(yottadb.NOTTP, nil, subs[0], subs[1:])
	return
}

// Set
func Set(glvn string, value string) (err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return
	}
	return yottadb.SetValE(yottadb.NOTTP, nil, value, subs[0], subs[1:])
}

// SetArg decompses an argument for the set command 's glvn=arg'
func SetArg(arg string) (glvn string, value string, err error) {
	if !strings.ContainsRune(arg, '=') {
		value, err = G(arg)
		if err != nil {
			return arg, "", errors.New("contains no '='")
		}
		arg += "=" + value
	}
	var parts []string
	k := strings.IndexAny(arg, "/(=")
	if arg[k:k+1] == "=" {
		parts = strings.SplitN(arg, "=", 2)
	} else {
		var e error
		if glvn[k:k+1] == "/" {
			parts, e = splits(arg)
		} else {
			parts, e = splitb(arg)
		}
		if e != nil {
			return "", "", e
		}
	}
	glvn = strings.TrimSpace(parts[0])
	glvn, _, err = QS(glvn)
	if err != nil {
		return "", "", err
	}
	value, _, _ = qutil.Nature(parts[1])
	return glvn, value, nil
}

// Next = $Next
func Next(glvn string) (nglvn, value string, err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return
	}
	z, e := yottadb.SubNextE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
	}
	subs[len(subs)-1] = z
	return UnQS(subs), z, nil
}

// Prev = $Next(-1)
func Prev(glvn string) (pglvn string, value string, err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return
	}
	z, e := yottadb.SubPrevE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
	}
	subs[len(subs)-1] = z
	return UnQS(subs), z, nil
}

func Right(glvn string) (rglvn string, err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return glvn, err
	}
	if len(subs) > 1 && subs[len(subs)-1] == "" {
		return glvn, err
	}
	subs = append(subs, "")
	return UnQS(subs), nil
}

func Left(glvn string) (rglvn string, err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return glvn, err
	}
	if len(subs) < 2 {
		return glvn, err
	}
	subs = subs[:len(subs)-1]
	return UnQS(subs), nil
}

func KillN(glvn string) (err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return err
	}
	err = yottadb.DeleteE(yottadb.NOTTP, nil, yottadb.YDB_DEL_NODE, subs[0], subs[1:])
	return err
}

func KillT(glvn string) (err error) {
	_, subs, err := QS(glvn)
	if err != nil {
		return err
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
	glvn = qutil.Escape(glvn)
	if strings.HasPrefix(glvn, "=") {
		return nil, errors.New("contains no reference part")
	}
	offset := 0
	for {
		k := strings.IndexByte(glvn[offset:], '=')
		if k == -1 {
			value, err := G(glvn)
			if err != nil {
				return nil, errors.New("contains no '='")
			}
			glvn += "=" + qutil.Escape(value)
			continue
		}
		if k == 0 {
			return nil, errors.New("contains no reference part")
		}
		before := glvn[:offset+k]
		if strings.Count(before, `"`)%2 == 0 {
			parts = append(parts, qutil.Unescape(before), qutil.Unescape(glvn[k+offset+1:]))
			return parts, nil
		}
		offset = offset + k + 1
	}
}

func Exit() {
	yottadb.Exit()
}

func Exec(text string) error {
	fmtable := "/home/rphilips/.yottadb/yottadbaccess.ci"
	envvarSave := make(map[string]string)
	qutil.SaveEnvvars(&envvarSave, "yottadb_ci", "yottadb_routines")
	os.Setenv("yottadb_ci", fmtable)
	out := Space
	_, err := yottadb.CallMT(yottadb.NOTTP, nil, 0, "xecute", text, &out)
	qutil.RestoreEnvvars(&envvarSave, "yottadb_ci", "yottadb_routines")
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
