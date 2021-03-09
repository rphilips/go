package lfile

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	qmumps "brocade.be/base/mumps"
	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// Lgcode stelt een specifieke lgcode voor
type Lgcode struct {
	ID       string `json:"id"`       // Identificatie
	N        string `json:"dut"`      // Nederlands
	E        string `json:"eng"`      // Engels
	F        string `json:"fre"`      // Frans
	D        string `json:"ger"`      // German
	U        string `json:"unv"`      // Universeel
	Alias    string `json:"alias"`    // Alias
	Nature   string `json:"nature"`   // Nature: string (default) | html | markdown
	Encoding string `json:"encoding"` // Encoding: UTF-8 (default) | xml
	Source   string `json:"source"`   // Editfile
	Line     string `json:"-"`        // Lijnnummer
	Version  string `json:"-"`        // Version
}

// *Lgcode moet de Object interface ondersteunen.
// Hierna volgen alle methodes die hiervoor nodig zijn.

// String
func (lgcode *Lgcode) String() string {
	return "l4_" + lgcode.Name()
}

// Name of lgcode
func (lgcode *Lgcode) Name() string {
	return lgcode.ID
}

// SetName of lgcode
func (lgcode *Lgcode) SetName(id string) {
	lgcode.ID = id
}

// Type of lgcode
func (lgcode *Lgcode) Type() string {
	return "l4"
}

// Release of lgcode
func (lgcode *Lgcode) Release() string {
	return lgcode.Version
}

// SetRelease of lgcode
func (lgcode *Lgcode) SetRelease(version string) {
	lgcode.Version = version
}

// EditFile of lgcode
func (lgcode *Lgcode) EditFile() string {
	return lgcode.Source
}

// SetEditFile of lgcode
func (lgcode *Lgcode) SetEditFile(source string) {
	lgcode.Source = source
}

// Lineno of macro
func (lgcode *Lgcode) Lineno() string {
	return lgcode.Line
}

// SetLineno of macro
func (lgcode *Lgcode) SetLineno(lineno string) {
	lgcode.Line = lineno
}

// MarshalJSON of lgcode
func (lgcode *Lgcode) MarshalJSON() ([]byte, error) {
	return json.Marshal(*lgcode)
}

// Unmarshal of lgcode
func (lgcode *Lgcode) Unmarshal(blob []byte) error {
	return json.Unmarshal(blob, lgcode)
}

// Loads from blob
func (lgcode *Lgcode) Loads(blob []byte) error {
	fname := lgcode.EditFile()
	blob = bytes.TrimSpace(qutil.Decomment(blob).Bytes())
	x, ep := Parse(fname, blob, Entrypoint("Lgcode"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"lgcode.parse"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{msg},
		}
		return err
	}
	if x == nil {
		return nil
	}
	*lgcode = *(x.(*Lgcode))
	lgcode.Source = fname
	return nil
}

// Mumps genereert een M structuur
func (lgcode *Lgcode) Mumps(batchid string) (mumps qmumps.MUMPS) {
	alias := lgcode.Alias
	lgco := lgcode.AliasResolve()
	alias = lgcode.Alias
	if alias != "" {
		lgco.ID = alias
		qobject.Fetch(&lgco)
	}
	replaceit(&lgco)

	m := qmumps.M{
		Subs:   []string{"ZA"},
		Action: "kill",
	}

	mumps = qmumps.MUMPS{m}

	m = qmumps.M{
		Subs:   []string{"node"},
		Value:  "todo",
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"batchid"},
		Value:  batchid,
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "type"},
		Value:  "lgcode",
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "id"},
		Value:  lgcode.ID,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "source"},
		Value:  lgcode.EditFile(),
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "dut"},
		Value:  lgco.N,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "eng"},
		Value:  lgco.E,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "fre"},
		Value:  lgco.F,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "ger"},
		Value:  lgco.D,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "unv"},
		Value:  lgco.U,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "nature"},
		Value:  lgcode.Nature,
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "data", "alias"},
		Value:  lgcode.Alias,
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Value:  "s recno=$I(^ZQTECH(node,batchid))",
		Action: "exec",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Value:  "m ^ZQTECH(node,batchid,recno)=ZA",
		Action: "exec",
	}

	mumps = append(mumps, m)

	return
}

func replaceit(lgcode *Lgcode) {
	r := lgcode.Release()
	lgcode.N = replaceInString(lgcode.N, "N", r)
	lgcode.E = replaceInString(lgcode.E, "E", r)
	lgcode.D = replaceInString(lgcode.D, "D", r)
	lgcode.F = replaceInString(lgcode.F, "F", r)
	lgcode.U = replaceInString(lgcode.U, "U", r)
}

func replaceInString(s string, lg string, r string) string {
	if !strings.Contains(s, "l4_") {
		return s
	}
	parts := strings.SplitN(s, "l4_", -1)
	pieces := make([]string, 0)

	odd := true
	for _, part := range parts {
		odd = !odd
		if !odd {
			pieces = append(pieces, part)
			continue
		}
		rest := ""
		for _, ch := range part {
			if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", ch) {
				break
			}
			rest += string(ch)
		}
		if rest == "" {
			pieces = append(pieces, "l4_"+part)
			continue
		}
		lgco := Lgcode{
			ID:      rest,
			Version: r,
		}
		err := qobject.Fetch(&lgco)
		if err != nil {
			pieces = append(pieces, "l4_"+part)
			continue
		}
		lgco = (&lgco).AliasResolve()

		text := ""
		switch lg {
		case "N":
			text = lgco.N
		case "E":
			text = lgco.E
		case "D":
			text = lgco.D
		case "F":
			text = lgco.F
		case "U":
			text = lgco.U
		}

		text = replaceInString(text, lg, r)
		pieces = append(pieces, text)
		pieces = append(pieces, strings.TrimPrefix(part, rest))
	}
	return strings.Join(pieces, "")
}

// AliasResolve geeft eem Lgcode terug die uiteeindelijk geen alias bevat
func (lgcode *Lgcode) AliasResolve() Lgcode {
	alias := lgcode.Alias
	if alias == "" {
		return *lgcode
	}
	lgco := *lgcode
	lgco.ID = alias
	qobject.Fetch(&lgco)
	return (&lgco).AliasResolve()
}

// Deps lijst de objecten op waarvan dit object aghankelijk is
func (lgcode *Lgcode) Deps() []byte {
	x := lgcode.Alias
	if x != "" {
		if !strings.HasPrefix(x, "l4_") {
			x = "l4_" + x
		}
	}
	x += "\n" + lgcode.N + "\n" + lgcode.E + "\n" + lgcode.D + "\n" + lgcode.F + "\n" + lgcode.U
	return []byte(x)
}

// Lint lgcode
func (lgcode *Lgcode) Lint() (errslice qerror.ErrorSlice) {

	testempty := false

	fname := lgcode.EditFile()
	lineno, _ := strconv.Atoi(lgcode.Lineno())
	name := lgcode.String()
	isscope := false
	if strings.Count(name, ".") == 2 && strings.HasSuffix(name, ".scope") {
		isscope = true
	}

	id := lgcode.ID
	x := lgcode.N + lgcode.E + lgcode.D + lgcode.F + lgcode.U + lgcode.Alias + lgcode.Nature
	if testempty && !isscope && strings.TrimSpace(x) == "" && lgcode.Nature != "empty" {
		err := &qerror.QError{
			Ref:    []string{"lgcode.lint.empty"},
			File:   fname,
			Lineno: lineno,
			Object: name,
			Type:   "Error",
			Msg:    []string{"`" + id + "` has no translation or alias"},
		}
		errslice = append(errslice, err)
		return errslice
	}

	nature := lgcode.Nature
	encoding := lgcode.Encoding
	if nature != "" && nature != "markdown" && nature != "rest" && nature != "mumps" && nature != "empty" {
		err := &qerror.QError{
			Ref:    []string{"lgcode.lint.nature"},
			File:   fname,
			Lineno: lineno,
			Object: name,
			Type:   "Error",
			Msg:    []string{"`" + id + "` has bad nature"},
		}
		errslice = append(errslice, err)
	}
	if encoding != "" && encoding != "xml" {
		err := &qerror.QError{
			Ref:    []string{"lgcode.lint.nature"},
			File:   fname,
			Lineno: lineno,
			Object: name,
			Type:   "Error",
			Msg:    []string{"`" + id + "` has bad nature"},
		}
		errslice = append(errslice, err)
	}

	// check on alias
	lreg := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*\.[a-zA-Z][a-zA-Z0-9]*$`)
	alias := lgcode.Alias
	if alias != "" {
		y := lgcode.N + lgcode.E + lgcode.D + lgcode.F + lgcode.U + lgcode.Encoding
		if strings.TrimSpace(y) != "" {
			err := &qerror.QError{
				Ref:    []string{"lgcode.lint.alias.nonempty"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"`" + id + "` is an alias and should not have other information"},
			}
			errslice = append(errslice, err)
		}

		if isscope {
			err := &qerror.QError{
				Ref:    []string{"lgcode.lint.alias.scope1"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"`" + id + "`: only aliases of type ns.text are allowed"},
			}
			errslice = append(errslice, err)
		}
		if isscope {
			err := &qerror.QError{
				Ref:    []string{"lgcode.lint.alias.scope2"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"`" + id + "`: only textfragments of type ns.text should have an alias"},
			}
			errslice = append(errslice, err)
		}
		matched := lreg.MatchString(x)
		if !matched {
			err := &qerror.QError{
				Ref:    []string{"lgcode.lint.alias.form"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"`" + id + "` refers to an alias which is not of a suitable form"},
			}
			errslice = append(errslice, err)
		}
	}
	if alias == "" {
		y := lgcode.N + lgcode.E + lgcode.D + lgcode.F + lgcode.U
		if testempty && strings.TrimSpace(y) == "" {
			err := &qerror.QError{
				Ref:    []string{"lgcode.lint.notalias.empty"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"`" + id + "` should have translations"},
			}
			errslice = append(errslice, err)
		}
	}
	if len(errslice) == 0 {
		return nil
	}
	return errslice

}

// Format ...
func (lgcode *Lgcode) Format() string {

	// header
	lines := []string{"lgcode " + lgcode.ID + ":"}
	sep1 := "«"
	sep2 := "»"
	x := lgcode.N + lgcode.E + lgcode.D + lgcode.F + lgcode.U
	if strings.ContainsAny(x, "«»") {
		sep1 = "⟦"
		sep2 = "⟧"
	}
	ok := false
	if lgcode.N != "" {
		lines = append(lines, "    N: "+sep1+lgcode.N+sep2)
		ok = true
	}
	if lgcode.E != "" {
		lines = append(lines, "    E: "+sep1+lgcode.E+sep2)
		ok = true
	}
	if lgcode.F != "" {
		lines = append(lines, "    F: "+sep1+lgcode.F+sep2)
		ok = true
	}
	if lgcode.D != "" {
		lines = append(lines, "    D: "+sep1+lgcode.D+sep2)
		ok = true
	}
	if lgcode.U != "" {
		lines = append(lines, "    U: "+sep1+lgcode.U+sep2)
		ok = true
	}
	if lgcode.Alias != "" {
		lines = append(lines, "    Alias: "+lgcode.Alias)
		ok = true
	}
	if !ok {
		lines = append(lines, "    N: "+sep1+lgcode.N+sep2)
		ok = true
	}
	if lgcode.Nature != "" {
		lines = append(lines, "    Nature: "+lgcode.Nature)
	}
	if lgcode.Encoding != "" {
		lines = append(lines, "    Encoding: "+lgcode.Encoding)
	}
	return strings.Join(lines, "\n")
}

// Replacer berekent de tekst die moet worden gebruikt bij de lgcode
func (lgcode *Lgcode) Replacer(env map[string]string, original string) string {
	if !strings.HasPrefix(original, "l4_") {
		return original
	}
	if strings.Count(original, "_") != 2 {
		return original
	}
	parts := strings.SplitN(original, "_", 3)
	algo := parts[1]
	if strings.IndexAny(algo, "NEFDU") != 0 {
		return original
	}
	suffix := algo[1:]
	lg := rune(algo[0])
	if suffix != "" && suffix != "php" && suffix != "py" && suffix != "js" {
		return original
	}
	data := ""
	switch lg {
	case 'N':
		data = lgcode.N
	case 'E':
		data = lgcode.E
	case 'D':
		data = lgcode.D
	case 'F':
		data = lgcode.F
	case 'U':
		data = lgcode.U

	}
	return qutil.ApplyAlgo(data, suffix)
}
