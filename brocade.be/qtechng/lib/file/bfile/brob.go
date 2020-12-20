package bfile

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	qmumps "brocade.be/base/mumps"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// Brob staat voor een Brocade macro
type Brob struct {
	ID      string   `json:"id"`     // Identificatie
	Ty      string   `json:"type"`   // Type
	Source  string   `json:"source"` // Editfile
	Body    []*Field `json:"body"`   // Body
	Line    string   `json:"-"`      // Lijnnummer
	Version string   `json:"-"`      // Version
}

// Field models a $-veld
type Field struct {
	key     string
	value   string
	attribs []*Duo
}

// Duo models a $$-attribute
type Duo struct {
	key   string
	value string
}

// Map verzamelt de attributen in een map
func (brob *Brob) Map(key string, count int) map[string]string {
	m := make(map[string]string)

	if key == "" {
		cnt := make(map[string]int)
		for _, f := range brob.Body {
			if cnt[f.key] == count-1 {
				m[f.key] = f.value
			}
			cnt[f.key]++
		}
		return m
	}
	cnt := 0
	for _, f := range brob.Body {
		if f.key != key {
			continue
		}
		cnt++
		if cnt != count {
			continue
		}
		for _, a := range f.attribs {
			m[a.key] = a.value
		}
		return m
	}
	return nil
}

// List verzamelt de attributen in een map
func (brob *Brob) List() [][3]string {
	m := make([][3]string, 0)
	for _, f := range brob.Body {
		m = append(m, [3]string{f.key, "", f.value})
		for _, a := range f.attribs {
			m = append(m, [3]string{f.key, a.key, a.value})
		}
	}
	return m
}

//
// *Brob moet de Object interface ondersteunen.
// Hierna volgen alle methodes die hiervoor nodig zijn.

// String
func (brob *Brob) String() string {
	return "b4_" + brob.Name()
}

// Name of brob
func (brob *Brob) Name() string {
	return brob.ID
}

// SetName of macro
func (brob *Brob) SetName(id string) {
	brob.ID = id
}

// Type of macro
func (brob *Brob) Type() string {
	return "b4_" + brob.Ty
}

// Release of macro
func (brob *Brob) Release() string {
	return brob.Version
}

// SetRelease of macro
func (brob *Brob) SetRelease(version string) {
	brob.Version = version
}

// EditFile of macro
func (brob *Brob) EditFile() string {
	return brob.Source
}

// SetEditFile of macro
func (brob *Brob) SetEditFile(source string) {
	brob.Source = source
}

// Lineno of macro
func (brob *Brob) Lineno() string {
	return brob.Line
}

// SetLineno of macro
func (brob *Brob) SetLineno(lineno string) {
	brob.Line = lineno
}

// Marshal of macro
func (brob *Brob) Marshal() ([]byte, error) {
	return json.MarshalIndent(brob, "", "    ")
}

// Unmarshal of macro
func (brob *Brob) Unmarshal(blob []byte) error {
	return json.Unmarshal(blob, brob)
}

// Loads from blob
func (brob *Brob) Loads(blob []byte) error {
	fname := brob.EditFile()
	blob = bytes.TrimSpace(qutil.Decomment(blob, "/").Bytes())

	x, ep := Parse(fname, blob, Entrypoint("Brob"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"brob.parse"},
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
	*brob = *(x.(*Brob))
	brob.Source = fname
	return nil
}

// Deps genereert de string waarin de objecten zitten
func (brob *Brob) Deps() []byte {
	buffer := new(bytes.Buffer)
	for _, l := range brob.List() {
		buffer.WriteString(l[2])
		buffer.WriteString("\n")
	}
	return buffer.Bytes()
}

// Lint macro
func (brob *Brob) Lint() (errslice qerror.ErrorSlice) {
	if len(errslice) == 0 {
		return nil
	}
	return errslice
}

// Format ...
func (brob *Brob) Format() string {
	lines := make([]string, 0)
	lines = append(lines, brob.Ty+" "+brob.ID+":")
	for _, field := range brob.Body {
		value := qutil.Embrace(field.value)
		lines = append(lines, "    $"+field.key+": "+value)
		for _, duo := range field.attribs {
			value := qutil.Embrace(duo.value)
			lines = append(lines, "        $$"+duo.key+": "+value)
		}
	}
	return strings.Join(lines, "\n")
}

// Replacer berekent de tekst die moet worden gebruikt bij de macro
func (brob *Brob) Replacer(env map[string]string, original string) string {
	return ""
}

// Mumps genereert een M structuur
func (brob *Brob) Mumps(batchid string) (mumps qmumps.MUMPS) {

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
		Value:  brob.Ty,
		Action: "set",
	}

	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "id"},
		Value:  brob.ID,
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "source"},
		Value:  brob.EditFile(),
		Action: "set",
	}
	mumps = append(mumps, m)

	// specifieke data

	l := brob.List()
	cnt := make(map[string]int)

	for _, one := range l {
		field := one[0]
		key := one[1]
		value := one[2]
		if key == "" {
			key = " "
			cnt[field+" "]++
		}
		nr1 := cnt[field+" "]
		snr1 := strconv.Itoa(nr1)

		cnt[field+" "+snr1+" "+key]++
		nr2 := cnt[field+" "+snr1+" "+key]
		snr2 := strconv.Itoa(nr2)

		m = qmumps.M{
			Subs:   []string{"field"},
			Value:  field,
			Action: "set",
		}
		mumps = append(mumps, m)

		m = qmumps.M{
			Subs:   []string{"nr1"},
			Value:  snr1,
			Action: "set",
		}
		mumps = append(mumps, m)

		m = qmumps.M{
			Subs:   []string{"key"},
			Value:  key,
			Action: "set",
		}
		mumps = append(mumps, m)

		m = qmumps.M{
			Subs:   []string{"nr2"},
			Value:  snr2,
			Action: "set",
		}
		mumps = append(mumps, m)

		m = qmumps.M{
			Subs:   []string{"v"},
			Value:  value,
			Action: "set",
		}
		mumps = append(mumps, m)

		m = qmumps.M{
			Value:  `s ZA("data",field,nr1,key,nr2)=v`,
			Action: "exec",
		}
		mumps = append(mumps, m)
	}

	// finalising

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
