package xfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	qmumps "brocade.be/base/mumps"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

var rverb = regexp.MustCompile(`x4_varruntime\(|x4_varcoderuntime\(|x4_parconstant\(|x4_parcode\(|x4_varcode\(|x4_vararray\(|x4_format\(|x4_exec\(|x4_if\(|x4_select\(|x4_lookupinitscreen\(|x4_lookupinitformat\(`)
var rdequote = regexp.MustCompile(`"(\$\$.*\))"`)

// Widget staat voor een Brocade macro
type Widget struct {
	ID      string `json:"id"`     // Identificatie
	Source  string `json:"source"` // Editfile
	Body    string `json:"body"`   // Body
	Line    string `json:"-"`      // Lijnnummer
	Version string `json:"-"`      // Version
	Text    string `json:"text"`   // Text
}

// String
func (widget *Widget) String() string {
	return widget.ID
}

// Name of widget
func (widget *Widget) Name() string {
	return widget.ID
}

// SetName of macro
func (widget *Widget) SetName(id string) {
	widget.ID = id
}

// Type of macro
func (widget *Widget) Type() string {
	x := widget.ID
	return strings.SplitN(x, " ", 2)[0]
}

// Release of macro
func (widget *Widget) Release() string {
	return widget.Version
}

// SetRelease of macro
func (widget *Widget) SetRelease(version string) {
	widget.Version = version
}

// Text of widget
func (widget *Widget) SetText(text string) {
	widget.Text = text
}

// EditFile of macro
func (widget *Widget) EditFile() string {
	return widget.Source
}

// SetEditFile of macro
func (widget *Widget) SetEditFile(source string) {
	widget.Source = source
}

// Lineno of macro
func (widget *Widget) Lineno() string {
	return widget.Line
}

// SetLineno of macro
func (widget *Widget) SetLineno(lineno string) {
	widget.Line = lineno
}

// MarshalJSON of macro
func (widget *Widget) MarshalJSON() ([]byte, error) {
	return json.Marshal(*widget)
}

// Unmarshal of macro
func (widget *Widget) Unmarshal(blob []byte) error {
	return json.Unmarshal(blob, widget)
}

// Loads from blob
func (widget *Widget) Loads(blob []byte) error {
	fname := widget.EditFile()
	//blob = bytes.TrimSpace(qutil.Decomment(blob).Bytes())
	x, ep := Parse(fname, blob, Entrypoint("Widget"))

	if ep != nil {
		msg := qutil.ExtractMsg(ep.Error(), fname)
		err := &qerror.QError{
			Ref:    []string{"widget.parse"},
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
	*widget = *(x.(*Widget))
	widget.Source = fname
	return nil
}

// Deps genereert de string waarin de objecten zitten
func (widget *Widget) Deps() []byte {
	buffer := new(bytes.Buffer)
	buffer.WriteString(widget.Body)
	return buffer.Bytes()
}

// Lint macro
func (widget *Widget) Lint() (errslice qerror.ErrorSlice) {
	if len(errslice) == 0 {
		return nil
	}
	return errslice
}

// Format ...
func (widget *Widget) Format() string {

	return ""
}

// Replacer berekent de tekst die moet worden gebruikt bij de macro
func (widget *Widget) Replacer(env map[string]string, original string) string {
	return ""
}

// X4 represnts a x4 construct
type X4 struct {
	mode  string
	text  string
	label string
	ask   int
	give  int
}

// Resolve maakt een lijst aan met M opdrachten
func (widget *Widget) Resolve() ([]string, error) {
	ty := widget.Type()
	if ty == "text" {
		return nil, nil
	}

	body := widget.Body

	if strings.HasPrefix(widget.ID, "format $") || strings.HasPrefix(widget.ID, "format @") {
		return []string{"-:" + body}, nil
	}

	if body == "" {
		return []string{"-:"}, nil
	}

	lineno, _ := strconv.Atoi(widget.Lineno())
	x4all := make([]X4, 0)
	for body != "" {
		x4s, s, e := x4extract(ty, body, "")
		body = s
		if e != "" {
			err := &qerror.QError{
				Ref:     []string{"xfile.parse.extract"},
				Version: widget.Release(),
				QPath:   widget.EditFile(),
				Lineno:  lineno,
				Object:  widget.ID,
				Msg:     []string{e},
			}

			return nil, err
		}
		if len(x4s) == 0 {
			continue
		}
		// if len(x4all) != 0 && x4all[len(x4all)-1].mode == "I" {
		// 	buf0 := buf[0]
		// 	if buf0.mode == "-" {
		// 		k := strings.Index(buf0.text, "\n")
		// 		if k != -1 {
		// 			buf[0].text = buf0.text[k+1:]
		// 		}
		// 	}
		// }
		x4all = append(x4all, x4s...)
	}

	ilabels := make([]int, 0)

	for i, x4def := range x4all {
		if x4def.label == "" {
			continue
		}
		ilabels = append(ilabels, i)
	}

	if len(ilabels) == 0 {
		result := packX4(x4all)
		return stringX4(result), nil
	} else {
		for i := 0; i < len(ilabels); i++ {
			j := ilabels[len(ilabels)-i-1]
			x4all, _ = findLabel(x4all, j)
		}
	}

	for i, x4def := range x4all {
		label := x4def.label
		if label == "" {
			continue
		}
		ask := x4def.ask
		found := false
		for j := i + 1; j < len(x4all); j++ {
			if x4all[j].mode != "L" {
				continue
			}
			if x4all[j].give != ask {
				continue
			}
			found = true
			x4all[i].text = strconv.Itoa(j+1) + x4def.text
			break
		}
		if !found {
			err := &qerror.QError{
				Ref:     []string{"xfile.parse.label"},
				Version: widget.Release(),
				QPath:   widget.EditFile(),
				Lineno:  lineno,
				Object:  widget.ID,
				Msg:     []string{"label `" + label + "` not found"},
			}

			return nil, err
		}
	}
	for i := range x4all {
		if x4all[i].mode != "L" {
			continue
		}
		x4all[i].mode = "-"
		x4all[i].text = ""
	}
	//result := packX4(x4all)

	return stringX4(x4all), nil
}

func stringX4(x4s []X4) []string {
	b := make([]string, len(x4s))
	for i, x4 := range x4s {
		b[i] = x4.mode + ":" + x4.text
	}
	return b
}

func packX4(x4s []X4) (result []X4) {
	result = make([]X4, 0)
	offset := 0
	prev := false
	for _, x4 := range x4s {
		if offset != 0 && x4.mode == "I" {
			text := x4.text
			parts := strings.SplitN(text, ":", 1)
			off, _ := strconv.Atoi(parts[0])
			parts[0] = strconv.Itoa(off + offset)
			x4.text = strings.Join(parts, ":")
			continue
		}
		if x4.mode != "-" {
			result = append(result, x4)
			prev = false
			continue
		}
		if prev {
			result[len(result)-1].text += x4.text
			offset++
			continue
		}
		prev = true
		result = append(result, x4)
	}
	return
}

func findLabel(buffer []X4, index int) ([]X4, bool) {
	ok := false
	label := buffer[index].label
	if label == "" {
		return buffer, true
	}
	for i := index + 1; i < len(buffer); i++ {
		x4def := buffer[i]
		mode := x4def.mode
		if mode == "L" {
			continue
		}
		text := x4def.text

		if !strings.Contains(text, label) {
			continue
		}
		parts := strings.SplitN(text, label, -1)
		here := -1
		for z := 0; z < len(parts)-1; z++ {

			k := strings.LastIndex(parts[z], "\n")
			if k == -1 && z != 0 {
				continue
			}
			before := parts[z][k+1:]
			if strings.TrimSpace(before) != "" {
				continue
			}
			after := parts[1+z]
			if after != "" && strings.TrimLeft(after, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_") != after {
				continue
			}
			ok = true
			here = z
			break
		}
		if !ok {
			continue
		}
		buffer[index].ask = index + 1
		x := append([]X4{}, buffer[:i]...)
		before := strings.Join(parts[:here+1], label)
		after := strings.Join(parts[here+1:], label)
		xi := buffer[i]
		xi.text = before
		x = append(x, xi)
		xii := X4{
			mode: "L",
			text: label,
			give: index + 1,
		}
		x = append(x, xii)
		if after != "" {
			xiii := X4{
				mode: "-",
				text: after,
			}
			x = append(x, xiii)
		}
		x = append(x, buffer[i+1:]...)
		return x, ok
	}
	return buffer, ok
}

func x4extract(ty string, body string, hull string) (x4s []X4, rest string, err string) {
	x4s = make([]X4, 0)
	here := rverb.FindStringIndex(body)
	if here == nil {
		if body != "" {
			x4s = append(x4s, X4{
				mode: "-",
				text: body,
			})
		}
		rest = ""
		return
	}

	k1 := here[0]
	k2 := here[1] - 1
	verb := body[k1+3 : k2]
	rest = body[k2:]
	if !strings.ContainsRune(rest, ')') {
		x4s = append(x4s, X4{
			mode: "-",
			text: body,
		})
		rest = ""
		return
	}
	args, until, err := qutil.BuildArgs(rest)
	if err != "" {

		x4s = append(x4s, X4{
			mode: "-",
			text: body,
		})
		if len(rest) > 32 {
			rest = rest[:32]
		}
		err += " > " + verb + ": " + rest
		rest = ""
		return
	}
	for i, arg := range args {
		args[i] = strings.ReplaceAll(strings.ReplaceAll(arg, "«", ""), "»", "")
	}

	rest = rest[len(until):]

	var result X4
	switch verb {
	case "varruntime":
		result, err = x4varruntime(args, ty, hull)
	case "varcoderuntime":
		result, err = x4varcoderuntime(args, ty, hull)
	case "parconstant":
		result, err = x4parconstant(args, ty, hull)
	case "parcode":
		result, err = x4parcode(args, ty, hull)
	case "varcode":
		result, err = x4varcode(args, ty, hull)
	case "vararray":
		result, err = x4vararray(args, ty, hull)
	case "format":
		result, err = x4format(args, ty, hull)
	case "exec":
		result, err = x4exec(args, ty, hull)
	case "if":
		result, err = x4if(args, ty, hull)
	case "select":
		result, err = x4select(args, ty, hull)
	case "lookupinitscreen":
		result, err = x4lookupinitscreen(args, ty, hull)
	case "lookupinitformat":
		result, err = x4lookupinitformat(args, ty, hull)
	}

	if err != "" {
		x4s = append(x4s, X4{
			mode: "-",
			text: body,
		})
		rest = ""
		return
	}
	x4s = append(x4s, X4{
		mode: "-",
		text: body[:k1],
	})

	x4s = append(x4s, result)
	return
}

func mexpr(arg string, quote bool) string {
	if !strings.Contains(arg, "x4_") {
		if !quote {
			return arg
		}
		return `"` + strings.ReplaceAll(arg, `"`, `""`) + `"`
	}
	rest := arg
	x4all := make([]X4, 0)
	for {
		x4s, rest1, _ := x4extract("format", rest, "if")
		x4all = append(x4all, x4s...)
		if rest1 == "" || rest == rest1 {
			break
		}
		rest = rest1
	}
	result := ""
	t := ""
	for _, x4 := range x4all {
		if x4.mode == "-" && x4.text != "" {
			t += x4.text
			continue
		}
		if t != "" {

			if quote {
				t = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
			}
			if result != "" && quote {
				result += "_" + t
			} else {

				result += t
			}
			t = ""
		}
		if result != "" && quote {
			result += "_" + x4.text
		} else {
			result += x4.text
		}
	}
	if t != "" {
		if quote {
			t = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
		}
		if result != "" && quote {
			result += "_" + t
		} else {

			result += t
		}
		t = ""
	}
	return rdequote.ReplaceAllString(result, `$1`)
}

func x4varruntime(args []string, ty string, hull string) (result X4, err string) {
	if hull != "" {
		switch hull {
		case "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
		default:
			err = fmt.Sprintf("x4_varruntime cannot be embedded in `%s`", hull)
			return
		}
	}
	// check on arguments
	if len(args) > 2 {
		err = fmt.Sprintf("x4_varruntime cannot work with %d arguments", len(args))
		return
	}

	enc := ""
	if len(args) == 2 {
		check := args[1]
		if strings.Contains(check, "_") {
			check = strings.SplitN(check, "_", 2)[0]
		}
		switch check {
		case "raw", "html", "js", "url", "num", "date", "xml":
			enc = args[1]
		default:
			err = fmt.Sprintf("x4_varruntime cannot work with second argument `%s`", args[1])
			return
		}
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`w $$Runtime^uwwwscr(%s,"%s")`, dnull, enc)
	} else {
		result.text = fmt.Sprintf(`$$Runtime^uwwwscr(%s,"%s")`, dnull, enc)
	}

	return
}

func x4varcoderuntime(args []string, ty string, hull string) (result X4, err string) {
	if hull != "" {
		switch hull {
		case "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
		default:
			err = fmt.Sprintf("x4_varcoderuntime cannot be embedded in `%s`", hull)
			return
		}
	}
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_varcoderuntime cannot work with %d arguments", len(args))
		return
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`w $$RuntimeC^uwwwscr(%s,"")`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$RuntimeC^uwwwscr(%s,"")`, dnull)
	}
	return
}

func x4varcode(args []string, ty string, hull string) (result X4, err string) {
	if hull != "" {
		switch hull {
		case "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
		default:
			err = fmt.Sprintf("x4_varcode cannot be embedded in `%s`", hull)
			return
		}
	}
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_varcode cannot work with %d arguments", len(args))
		return
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`w $$TrlCode^uwwwscr(%s)`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$TrlCode^uwwwscr(%s,"")`, dnull)
	}

	return
}

func x4format(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_format cannot be used in other x4 constructions"
		return
	}
	// check on ty
	if ty == "format" {
		err = "x4_format does not work in formats"
	}
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_format cannot work with %d arguments", len(args))
		return
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	if dnull == "" {
		err = "x4_format does not work with empty argument"
		return
	}

	// result construction
	result.mode = "X"
	result.text = fmt.Sprintf(`d %%OpenArr^uwwwscr("",""),%%InitRow^uwwwscr("%s",""),%%AddPar^uwwwscr(""),%%ClosArr^uwwwscr d Format^uwwwscr("")`, dnull)

	return
}

func x4parconstant(args []string, ty string, hull string) (result X4, err string) {
	if hull != "" {
		switch hull {
		case "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
		default:
			err = fmt.Sprintf("x4_varcode cannot be embedded in `%s`", hull)
			return
		}
	}
	if ty == "screen" {
		err = "x4_parconstant does not work in screens"
	}
	// check on arguments
	if len(args) > 2 {
		err = fmt.Sprintf("x4_parconstant cannot work with %d arguments", len(args))
		return
	}

	switch hull {
	case "", "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
	default:
		err = fmt.Sprintf("x4_parconstant cannot work inside `%s`", hull)
		return
	}

	enc := ""
	if len(args) == 2 {
		check := args[1]
		if strings.Contains(check, "_") {
			check = strings.SplitN(check, "_", 2)[0]
		}
		switch check {
		case "raw", "html", "js", "url", "num", "date", "xml":
			enc = args[1]
		default:
			err = fmt.Sprintf("x4_parconstant cannot work with second argument `%s`", args[1])
			return
		}
	}
	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`w $$GetParVl^uwwwscr(PRvar,%s,"CONSTANT","%s")`, dnull, enc)
	} else {
		result.text = fmt.Sprintf(`$$GetParVl^uwwwscr(PRvar,%s,"CONSTANT","%s")`, dnull, enc)
	}
	return
}

func x4parcode(args []string, ty string, hull string) (result X4, err string) {
	if hull != "" {
		switch hull {
		case "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
		default:
			err = fmt.Sprintf("x4_parcode cannot be embedded in `%s`", hull)
			return
		}
	}
	if ty == "screen" {
		err = "x4_parcode does not work in screens"
	}
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_parcode cannot work with %d arguments", len(args))
		return
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`w $$GetParVl^uwwwscr(PRvar,%s,"CODE")`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$GetParVl^uwwwscr(PRvar,%s,"CODE")`, dnull)
	}
	return
}

func x4vararray(args []string, ty string, hull string) (result X4, err string) {
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_vararray cannot work with %d arguments", len(args))
		return
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	dnull = `"` + strings.ReplaceAll(dnull, `"`, `""`) + `"`

	// result construction
	if hull == "" {
		result.mode = "X"
		result.text = fmt.Sprintf(`d Format^uwwwscr(%s)`, dnull)
	}
	return
}
func x4exec(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_exec cannot be used in other x4 constructions"
		return
	}
	// check on arguments
	if len(args) != 1 {
		err = "x4_exec works with exactly 1 parameter"
		return
	}

	par := mexpr(args[0], false)

	result.mode = "X"
	result.text = par
	return
}

func x4if(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_if cannot be used in other x4 constructions"
		return
	}

	if len(args) != 2 {
		err = "x4_if works with exactly 2 parameters"
		return
	}
	label := strings.TrimSpace(args[0])
	if label == "" {
		err = "x4_if should have a non-emyty label"
	}
	arg1 := mexpr(strings.TrimSpace(args[1]), false)
	if strings.ContainsAny(arg1, "«»⟦⟧") {
		arg1 = strings.ReplaceAll(arg1, "«", "")
		arg1 = strings.ReplaceAll(arg1, "»", "")
		arg1 = strings.ReplaceAll(arg1, "⟦", "")
		arg1 = strings.ReplaceAll(arg1, "⟧", "")
	}
	result.mode = "I"
	result.text = ":" + arg1
	result.label = label

	return
}
func x4select(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_select cannot be used in other x4 constructions"
		return
	}
	if len(args) != 3 {
		err = "x4_select works with exactly 3 parameters"
		return
	}

	result.mode = "X"
	result.text = fmt.Sprintf(`w $s(%s:%s,1:%s)`, mexpr(args[2], false), mexpr(args[0], true), mexpr(args[1], true))
	return
}

func x4lookupinitscreen(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_lookupinitscreen cannot be used in other x4 constructions"
		return
	}
	if len(args) < 2 {
		err = "x4_lookupinitscreen works with at least 2 parameters"
		return
	}
	if len(args) > 5 {
		err = "x4_lookupinitscreen works with no more than 5 parameters"
		return
	}
	v0 := mexpr(args[0], true)
	v1 := mexpr(args[1], true)
	v2 := `""`
	v3 := `""`
	v4 := `""`
	switch len(args) {
	case 5:
		v2 = mexpr(args[2], true)
		v3 = mexpr(args[3], true)
		v4 = mexpr(args[4], true)
	case 4:
		v2 = mexpr(args[2], true)
		v3 = mexpr(args[3], true)
	case 3:
		v2 = mexpr(args[2], true)
	}
	result.mode = "X"
	result.text = fmt.Sprintf(`d %%Entry^uluwake(%s,%s,%s,$$Runtime^uwwwscr(%s),%s)`, v0, v1, v2, v3, v4)
	if v3 == `""` {
		result.text = fmt.Sprintf(`d %%Entry^uluwake(%s,%s,%s,%s,%s)`, v0, v1, v2, v3, v4)
	}
	return
}
func x4lookupinitformat(args []string, ty string, hull string) (result X4, err string) {
	// not write
	if hull != "" {
		err = "x4_lookupinitformat cannot be used in other x4 constructions"
		return
	}
	if ty == "screen" {
		err = "x4_lookupinitformat does not work in screens"
	}
	if len(args) < 2 {
		err = "x4_lookupinitformat works with at least 2 parameters"
		return
	}
	if len(args) > 5 {
		err = "x4_lookupinitformat works with no more than 5 parameters"
		return
	}

	v0 := args[0]
	v1 := args[1]
	v2 := ""
	v3 := ``
	v4 := ""
	switch len(args) {
	case 5:
		v2 = args[2]
		v3 = args[3]
		v4 = args[4]
	case 4:
		v2 = args[2]
		v3 = args[3]
	case 3:
		v2 = args[2]
	}
	v0 = `"` + strings.ReplaceAll(v0, `"`, `""`) + `"`
	v1 = `"` + strings.ReplaceAll(v1, `"`, `""`) + `"`
	v4 = `"` + strings.ReplaceAll(v4, `"`, `""`) + `"`
	result.mode = "X"
	result.text = fmt.Sprintf(`d %%Entry^uluwake(%s,%s,$$GetParVl^uwwwscr(PRvar,%s,"CONSTANT",""),$$GetParVl^uwwwscr(PRvar,%s,"CONSTANT",""),%s)`, v0, v1, v2, v3, v4)
	return
}

// Mumps genereert een M structuur
func (widget *Widget) Mumps(batchid string) (mumps qmumps.MUMPS, err error) {

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

	parts := strings.SplitN(widget.ID, " ", 2)

	m = qmumps.M{
		Subs:   []string{"ZA", "type"},
		Value:  parts[0],
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "id"},
		Value:  parts[1],
		Action: "set",
	}
	mumps = append(mumps, m)

	m = qmumps.M{
		Subs:   []string{"ZA", "source"},
		Value:  widget.EditFile(),
		Action: "set",
	}
	mumps = append(mumps, m)

	lines, err := widget.Resolve()

	for i, line := range lines {
		recno := strconv.Itoa(i + 1)
		m = qmumps.M{
			Subs:   []string{"ZA", "data", recno},
			Value:  line,
			Action: "set",
		}
		mumps = append(mumps, m)
	}

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
