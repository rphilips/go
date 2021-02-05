package ifile

import (
	"encoding/json"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// Include staat een include
type Include struct {
	ID      string `json:"id"`      // Identificatie
	Content string `json:"content"` // Waarde
	Source  string `json:"source"`  // Editfile
	Line    string `json:"-"`       // Lijnnummer
	Version string `json:"-"`       // Version
}

// *Include moet de Object interface ondersteunen.
// Hierna volgen alle methodes die hiervoor nodig zijn.

// String
func (include *Include) String() string {
	return "i4_" + include.Name()
}

// Name of include
func (include *Include) Name() string {
	return include.ID
}

// SetName of include
func (include *Include) SetName(id string) {
	include.ID = id
}

// Type of include
func (include *Include) Type() string {
	return "i4"
}

// Release of include
func (include *Include) Release() string {
	return include.Version
}

// SetRelease of include
func (include *Include) SetRelease(version string) {
	include.Version = version
}

// EditFile of include
func (include *Include) EditFile() string {
	return include.Source
}

// SetEditFile of include
func (include *Include) SetEditFile(source string) {
	include.Source = source
}

// Lineno of include
func (include *Include) Lineno() string {
	return include.Line
}

// SetLineno of include
func (include *Include) SetLineno(lineno string) {
	include.Line = lineno
}

// Marshal of include
func (include *Include) Marshal() ([]byte, error) {
	return json.MarshalIndent(include, "", "    ")
}

// MarshalJSON of include
func (include *Include) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(include, "", "    ")
}

// Unmarshal of include
func (include *Include) Unmarshal(blob []byte) error {
	return json.Unmarshal(blob, include)
}

// Deps lijst de objecten op waarvan dit object aghankelijk is
func (include *Include) Deps() []byte {
	return []byte(include.Content)
}

// Loads from blob
func (include *Include) Loads(blob []byte) error {
	fname := include.EditFile()
	x, ep := Parse(fname, blob, Entrypoint("Include"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"include.parse"},
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
	*include = *(x.(*Include))
	include.Source = fname
	return nil
}

// Lint include
func (include *Include) Lint() (errslice qerror.ErrorSlice) {
	return nil
}

// Format ...
func (include *Include) Format() string {

	// header
	header := "include " + include.ID + ":"

	content := include.Content
	a := "«"
	b := "»"
	if strings.ContainsAny(content, a+b) {
		a = "⟦"
		b = "⟧"
	}
	result := []string{header, a + content + b}
	return strings.Join(result, "\n")
}

// Replacer berekent de tekst die moet worden gebruikt bij de include
func (include Include) Replacer(env map[string]string, original string) string {
	return include.Content
}
