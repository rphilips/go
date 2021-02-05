package dfile

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// Param staat voor een parameter
type Param struct {
	ID      string `json:"name"`    // Naam
	Ref     string `json:"ref"`     // Reference
	Default string `json:"default"` // default value
	Doc     string `json:"doc"`     // Documentation
	Named   bool   `json:"named"`   // Named?
}

// Action for the macro
type Action struct {
	Binary []string `json:"binary"` // compiled form
	Unless bool     `json:"unless"` // binding between action guard
	Guard  []string `json:"guard"`  // Guard
}

// Macro staat voor een Brocade macro
type Macro struct {
	ID       string   `json:"id"`       // Identificatie
	Synopsis string   `json:"synopsis"` // Synopsis
	Params   []Param  `json:"params"`   // parameters
	Actions  []Action `json:"actions"`  // Synopsis
	Examples []string `json:"examples"` // Examples
	Source   string   `json:"source"`   // Editfile
	Line     string   `json:"-"`        // Lijnnummer
	Version  string   `json:"-"`        // Version
}

// LoadsGuard from blob
func LoadsGuard(blob []byte) ([]string, error) {

	fname := ""
	x, ep := Parse(fname, blob, Entrypoint("Guard"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"parse.guard.parse"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{msg},
		}
		return nil, err
	}
	if x == nil {
		return nil, nil
	}
	return x.([]string), nil
}

// *Macro moet de Object interface ondersteunen.
// Hierna volgen alle methodes die hiervoor nodig zijn.

// String
func (macro *Macro) String() string {
	return "m4_" + macro.Name()
}

// Name of macro
func (macro *Macro) Name() string {
	return macro.ID
}

// SetName of macro
func (macro *Macro) SetName(id string) {
	macro.ID = id
}

// Type of macro
func (macro *Macro) Type() string {
	return "m4"
}

// Release of macro
func (macro *Macro) Release() string {
	return macro.Version
}

// SetRelease of macro
func (macro *Macro) SetRelease(version string) {
	macro.Version = version
}

// EditFile of macro
func (macro *Macro) EditFile() string {
	return macro.Source
}

// SetEditFile of macro
func (macro *Macro) SetEditFile(source string) {
	macro.Source = source
}

// Lineno of macro
func (macro *Macro) Lineno() string {
	return macro.Line
}

// SetLineno of macro
func (macro *Macro) SetLineno(lineno string) {
	macro.Line = lineno
}

// Marshal of macro
func (macro *Macro) Marshal() ([]byte, error) {
	return json.MarshalIndent(macro, "", "    ")
}

// MarshalJSON of macro
func (macro *Macro) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(macro, "", "    ")
}

// Unmarshal of macro
func (macro *Macro) Unmarshal(blob []byte) error {
	return json.Unmarshal(blob, macro)
}

// Loads from blob
func (macro *Macro) Loads(blob []byte) error {
	fname := macro.EditFile()
	blob = bytes.TrimSpace(qutil.Decomment(blob, "/").Bytes())
	x, ep := Parse(fname, blob, Entrypoint("Macro"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"macro.parse"},
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
	*macro = *(x.(*Macro))
	macro.Source = fname
	return nil
}

// Deps lijst de objecten op waarvan dit object afhankelijk is
func (macro *Macro) Deps() []byte {
	x := ""
	for _, action := range macro.Actions {
		bin := action.Binary
		act := strings.Join(bin, "")
		guard := action.Guard
		g := infix(guard)
		a := "«"
		b := "»"
		if strings.ContainsAny(act, a+b) || strings.ContainsAny(g, a+b) {
			a = "⟦"
			b = "⟧"
		}
		if g == "" {
			x += "\n" + a + act + b
			continue
		}
		unless := "if"
		if action.Unless {
			unless = "unless"
		}
		x += "\n" + a + act + b + "\n" + unless + " " + a + g + b
	}
	return []byte(x)
}

// Lint macro
func (macro *Macro) Lint() (errslice qerror.ErrorSlice) {

	fname := macro.EditFile()
	lineno, _ := strconv.Atoi(macro.Lineno())
	name := macro.String()
	// check synopsis
	if strings.TrimSpace(macro.Synopsis) == "" {
		e := &qerror.QError{
			Ref:    []string{"macro.lint.synopsis"},
			File:   fname,
			Lineno: lineno,
			Object: name,
			Type:   "Error",
			Msg:    []string{"No synopsis for `" + macro.String() + "`"},
		}
		errslice = append(errslice, e)
	}

	// check parameters

	for _, param := range macro.Params {
		doc := strings.TrimSpace(param.Doc)
		if doc == "" {
			e := &qerror.QError{
				Ref:    []string{"macro.lint.paramcomment"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"No comment for `" + param.ID + "` in `" + macro.String() + "`"},
			}
			errslice = append(errslice, e)
		}
	}

	// reference
	first := ""
	for i, param := range macro.Params {
		if i == 0 {
			first = param.ID
		}
		if param.Ref == "" {
			continue
		}
		if i != 1 {
			e := &qerror.QError{
				Ref:    []string{"macro.lint.ref1"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"Only the second parameter can have a reference"},
			}
			errslice = append(errslice, e)
			continue
		}

		if param.Ref != first {
			e := &qerror.QError{
				Ref:    []string{"macro.lint.ref2"},
				File:   fname,
				Lineno: lineno,
				Object: name,
				Type:   "Error",
				Msg:    []string{"The reference should be to the first parameter `" + first + "`"},
			}
			errslice = append(errslice, e)
			continue
		}
	}
	if len(errslice) == 0 {
		return nil
	}
	return errslice
}

// Format ...
func (macro *Macro) Format() string {

	// header
	header := "macro " + macro.ID
	var prms []string
	for _, param := range macro.Params {
		prm := ""
		if param.Named {
			prm = "*"
		}
		prm += param.ID
		def := param.Default
		def2 := def
		if len(def2) > 1 && strings.HasPrefix(def2, `"`) && strings.HasSuffix(def2, `"`) {
			def2 = def2[1 : len(def2)-1]
		}
		switch {
		case strings.ContainsAny(def2, `"(), `):
			prm += "=«" + def + "»"
		case def == "":
		default:
			prm += "=" + def
		}
		prms = append(prms, prm)
	}
	if len(prms) != 0 {
		header += "(" + strings.Join(prms, ", ") + ")"
	}
	header += ":"
	result := []string{header}
	result = append(result, "    '''")
	value := macro.Synopsis

	prefix := "$synopsis"
	result = append(result, collect(value, prefix)...)
	k := -1
	for _, param := range macro.Params {
		id := param.ID
		if len(id) > k {
			k = len(id)
		}
	}
	for _, param := range macro.Params {
		id := param.ID
		plus := ""
		if len(id) < k {
			plus = strings.Repeat(" ", k-len(id))
		}
		prefix := param.ID + plus
		value := param.Doc
		result = append(result, collect(value, prefix)...)
	}
	examples := macro.Examples
	prefix = "$example"
	for _, value := range examples {
		result = append(result, collect(value, prefix)...)
	}

	result = append(result, "    '''")

	for _, action := range macro.Actions {
		bin := action.Binary
		act := strings.Join(bin, "")
		guard := action.Guard
		g := infix(guard)
		a := "«"
		b := "»"
		if strings.ContainsAny(act, a+b) || strings.ContainsAny(g, a+b) {
			a = "⟦"
			b = "⟧"
		}
		if g == "" {
			result = append(result, "    "+a+act+b)
			continue
		}
		unless := "if"
		if action.Unless {
			unless = "unless"
		}
		result = append(result, "    "+a+act+b+"\n        "+unless+" "+a+g+b)
	}
	return strings.Join(result, "\n")
}

// Replacer berekent de tekst die moet worden gebruikt bij de macro
func (macro *Macro) Replacer(env map[string]string, original string) string {
	act := Action{}
	for _, action := range macro.Actions {
		guard := action.Guard
		truth := Eval(guard, env)
		if action.Unless {
			truth = !truth
		}
		if truth {
			act = action
			break
		}
	}
	binary := act.Binary
	if len(binary) == 0 {
		return ""
	}
	params := macro.Params
	defo := make(map[string]string)
	for _, param := range params {
		defo[param.ID] = param.Default
	}
	even := false
	result := make([]string, 0)
	for _, bin := range binary {
		even = !even
		if even {
			result = append(result, bin)
			continue
		}
		value, ok := env[bin]
		if !ok {
			value, ok = defo[bin]
			if !ok {
				return original
			}
		}
		result = append(result, value)
	}
	return strings.Join(result, "")
}

// Args zoekt de argumenten bij een macro
func (macro *Macro) Args(original string) (args map[string]string, rest string, err error) {
	x := original
	if !strings.HasPrefix(x, "(") {
		return nil, original, nil
	}
	xargs, until, msg := qutil.BuildArgs(x)
	args = make(map[string]string)
	done := make(map[string]bool)
	for _, param := range macro.Params {
		args[param.ID] = param.Default
	}
	for i, x := range xargs {
		if msg != "" {
			break
		}
		if i >= len(macro.Params) {
			msg = "Too many arguments"
			continue
		}
		k := strings.Index(x, "=")
		if k == -1 && macro.Params[i].Named {
			msg = "Parameter `" + macro.Params[i].ID + "` should be named"
			continue
		}
		z := macro.Params[i].ID
		value := x
		if k != -1 {
			z = strings.TrimSpace(x[:k])
			value = x[k+1:]
		}
		_, ok := args[z]
		if !ok {
			z = "$" + z
			_, ok = args[z]
		}
		if ok {
			if done[z] {
				msg = "Parameter `" + z + "` occurs twice"
				continue
			}

			args[z] = qutil.CleanArg(value)
			done[z] = true
			continue
		}
		if !ok {
			msg = "Parameter `" + z + "` is not specified"
			continue
		}
	}

	if msg != "" {
		err = &qerror.QError{
			Ref:    []string{"parse.args.parse"},
			File:   macro.EditFile(),
			Lineno: -1,
			Type:   "Error",
			Msg:    []string{msg},
		}
	}

	return args, original[len(until):], err
}

////////////////// Tools

func binary(s string, params []Param) []string {
	if len(params) == 0 {
		return []string{s}
	}
	pars := make([]string, 0)
	for _, p := range params {
		pars = append(pars, p.ID)
	}
	sort.Slice(pars, func(i, j int) bool { return len(pars[i]) > len(pars[j]) })
	result := []string{s}

	for _, p := range pars {
		res := result[:]
		result = make([]string, 0)
		ok := false
		for _, s := range res {
			ok = !ok
			if ok {
				r := paramsplit(s, p)
				result = append(result, r...)
				continue
			}
			result = append(result, s)
		}
	}
	return result
}

func paramsplit(s string, p string) (result []string) {
	if !strings.Contains(s, p) {
		return []string{s}
	}
	parts := strings.SplitN(s, p, -1)
	inp := false
	for _, part := range parts {
		if inp {
			result = append(result, p)
		}
		result = append(result, part)
		inp = true
	}
	return
}

func infix(guard []string) string {
	if len(guard) == 0 {
		return ""
	}
	pfix := []string{}
	term := ""
	operand := ""
	value := ""
	stadium := 0
	for _, s := range guard {
		switch s {
		case "not", "and", "or":
			pfix = append(pfix, " "+s+" ")
			term = ""
			operand = ""
			value = ""
			stadium = 0
		case "true", "false":
			pfix = append(pfix, s)
			term = ""
			operand = ""
			value = ""
			stadium = 0
		default:
			if stadium == 0 {
				stadium = 1
				term = s
				continue
			}
			if stadium == 1 {
				stadium = 2
				value = s
				continue
			}
			operand = s
			value = strings.ReplaceAll(value, `"`, `""`)
			pfix = append(pfix, term+" "+operand+" \""+value+"\"")
			term = ""
			operand = ""
			value = ""
			stadium = 0
		}
	}
	stack := []string{}
	st := -1
	for _, p := range pfix {
		switch p {
		case " not ":
			stack[st] = p + stack[st]
		case " and ", " or ":
			stack[st-1] = "(" + stack[st-1] + p + stack[st] + ")"
			st--
		case "false", "true":
			stack = append(stack, p)
			st++
		default:
			stack = append(stack, p)
			st++
		}
	}
	result := strings.TrimSpace(stack[0])

	return result

}

func collect(value string, prefix string) (result []string) {
	if value == "" {
		return []string{"    " + prefix + ":"}
	}
	prefix = "    " + prefix + ": "
	result = []string{prefix}
	lines := strings.SplitN(value, "\n", -1)
	first := true
	for _, line := range lines {

		if first {
			first = false
			line := strings.TrimSpace(line)
			result = []string{prefix + line}
			prefix = strings.Repeat(" ", len(prefix))
			continue
		}
		result = append(result, prefix+line)
	}
	return
}
