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

func N(glvn string) string {
	exec := `s %qzxy0=$NA(` + glvn + `)`
	r, err := Calc(exec, `%qzxy0`)
	if r == "" || err != nil {
		return glvn
	}
	return r
}

func EUnQS(subs []string) (glvn string) {
	switch len(subs) {
	case 0:
		return ""
	case 1:
		return subs[0]
	default:
		args := make([]string, len(subs))
		args[0] = subs[0]
		for i := 1; i < len(subs); i++ {
			args[i] = MakeArg(subs[i])
		}
		return UnQS(args)
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
func D(glvn string) (value int, err error) {
	glvn = N(glvn)
	subs := QS(glvn)
	argums := MArgs(subs)
	if len(argums) == 0 {
		return -1, nil
	}

	d, err := yottadb.DataE(yottadb.NOTTP, nil, argums[0], argums[1:])
	if err != nil {
		return -1, err
	}
	return int(d), nil
}

// G return $G() of reference, but returns an error
func G(glvn string, simple bool) (value string, err error) {
	subs := QS(glvn)
	if len(subs) == 0 {
		return "", errors.New("invalid reference")
	}
	if simple {
		argums := MArgs(subs)
		value, err = yottadb.ValE(yottadb.NOTTP, nil, argums[0], argums[1:])
		return
	}
	exec := "s %QwERTY=" + UnQS(subs)
	value, err = Calc(exec, "%QwERTY")
	return
}

// Set
func Set(glvn string, value string) (err error) {
	subs := QS(glvn)
	if len(subs) == 0 {
		return errors.New("invalid reference")
	}
	exec := "s " + UnQS(subs) + "=" + `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
	return Exec(exec)
}

// SetArg decompses an argument for the set command 's glvn=arg'
func SetArg(arg string) (glvn string, value string, err error) {
	if !strings.ContainsRune(arg, '=') {
		arg = Glvn(arg)
		value, err = G(arg, false)
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

func MakeArg(s string) string {
	if isNumber(s) {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// Next = $Next
func Next(glvn string) (nglvn, next string, err error) {
	subs := MArgs(QS(glvn))
	if len(subs) < 2 {
		return glvn, "", errors.New("cannot determine next")
	}
	z, e := yottadb.SubNextE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
		if e.Error() == "NODEEND" {
			e = nil
		}
		return glvn, "", e
	}
	subs[len(subs)-1] = z
	return EUnQS(subs), z, nil
}

// Prev = $Next(-1)
func Prev(glvn string) (pglvn string, value string, err error) {
	subs := MArgs(QS(glvn))
	if len(subs) < 2 {
		return glvn, "", errors.New("cannot determine prev")
	}
	z, e := yottadb.SubPrevE(yottadb.NOTTP, nil, subs[0], subs[1:])
	if e != nil {
		z = ""
		if e.Error() == "NODEEND" {
			e = nil
		}
		return glvn, "", e
	}
	subs[len(subs)-1] = z
	return EUnQS(subs), z, nil
}

func Right(glvn string) (rglvn string, err error) {
	subs := MArgs(QS(glvn))
	if len(subs) == 0 {
		return glvn, errors.New("cannot determine right")
	}
	if len(subs) > 1 && subs[len(subs)-1] == "" {
		return glvn, err
	}
	subs = append(subs, "")
	return EUnQS(subs), nil
}

func Left(glvn string) (rglvn string, err error) {
	subs := MArgs(QS(glvn))
	if len(subs) < 2 {
		return glvn, errors.New("cannot determine left")
	}
	subs = subs[:len(subs)-1]
	return EUnQS(subs), nil
}

func MArgs(subs []string) []string {
	if len(subs) == 0 {
		return subs
	}
	argums := make([]string, len(subs))
	for i := 0; i < len(subs); i++ {
		sub := subs[i]
		if i == 0 {
			argums[i] = sub
			continue
		}
		if !strings.HasPrefix(sub, `"`) {
			argums[i] = sub
			continue
		}
		if len(sub) < 2 {
			argums[i] = sub
			continue
		}
		if !strings.HasSuffix(sub, `"`) {
			argums[i] = sub
			continue
		}
		argums[i] = strings.ReplaceAll(sub[1:len(sub)-1], `""`, `"`)
	}
	return argums
}

func Kill(glvn string, tree bool) (err error) {
	glvn = N(glvn)
	subs := QS(glvn)
	if len(subs) == 0 {
		return errors.New("cannot determine node")
	}
	argums := MArgs(subs)
	kway := yottadb.YDB_DEL_TREE
	if !tree {
		kway = yottadb.YDB_DEL_NODE
	}
	err = yottadb.DeleteE(yottadb.NOTTP, nil, kway, argums[0], argums[1:])
	return err
}

func splits(glvn string) (parts []string, e error) {
	glvn = qutil.Escape(glvn)
	if !strings.ContainsRune(glvn, '=') {
		value, err := G(glvn, false)
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
			value, err := G(glvn, false)
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

func Calc(text string, lvar string) (value string, err error) {
	err = Exec(text)
	if err != nil {
		return "", err
	}
	value, err = yottadb.ValE(yottadb.NOTTP, nil, lvar, nil)
	return
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

type VarReport struct {
	Gloref string
	Value  string
	Err    error
}

func ZWR(gloref string, report chan VarReport, needle string) {
	rex := new(regexp.Regexp)
	var err error
	if needle != "" {
		rex, err = regexp.Compile(needle)
		if err != nil {
			rex = nil
		}
	}
	gloref = N(gloref)
	d, err := D(gloref)
	if d == 0 {
		err = fmt.Errorf("`%s` does not exist", gloref)
	}
	if err != nil {
		report <- VarReport{
			Gloref: "",
			Value:  "",
			Err:    err,
		}
		close(report)
		return
	}
	if needle == "" && (d == 1 || d == 11) {
		value, _ := G(gloref, true)
		report <- VarReport{
			Gloref: gloref,
			Value:  value,
			Err:    err,
		}
	}

	last := strings.TrimRight(gloref, "(),")
	for {
		exec := "s %nExt=$Q(" + gloref + ")"
		next, err := Calc(exec, "%nExt")

		if !strings.HasPrefix(next, last) {
			next = ""
		}
		if next != "" {
			x := strings.SplitN(next, last, 2)[0]
			if strings.IndexAny(x, "(),") == 0 {
				next = ""
			}

		}
		if err != nil || next == "" {
			report <- VarReport{
				Gloref: "",
				Value:  "",
				Err:    err,
			}
			close(report)
			return
		}
		gloref = next
		value, _ := G(gloref, true)
		if needle != "" {
			pair := gloref + "=" + value
			found := strings.Contains(pair, needle)
			if !found && rex != nil {
				found = rex.FindStringIndex(pair) != nil
			}
			if !found {
				continue
			}
		}

		report <- VarReport{
			Gloref: gloref,
			Value:  value,
			Err:    err,
		}
		if needle != "" {
			close(report)
			return
		}
	}

}
