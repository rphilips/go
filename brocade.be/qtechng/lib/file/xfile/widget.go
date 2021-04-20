package xfile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	qmumps "brocade.be/base/mumps"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// Widget staat voor een Brocade macro
type Widget struct {
	ID      string `json:"id"`     // Identificatie
	Source  string `json:"source"` // Editfile
	Body    string `json:"body"`   // Body
	Line    string `json:"-"`      // Lijnnummer
	Version string `json:"-"`      // Version
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
	return strings.SplitN(x, " ", 1)[0]
}

// Release of macro
func (widget *Widget) Release() string {
	return widget.Version
}

// SetRelease of macro
func (widget *Widget) SetRelease(version string) {
	widget.Version = version
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
}

// Resolve maakt een lijst aan met M opdrachten
func (widget *Widget) Resolve() ([]string, error) {
	ty := widget.Type()
	lineno, _ := strconv.Atoi(widget.Lineno())
	if ty == "text" {
		return nil, nil
	}

	body := widget.Body
	buffer := make([]X4, 0)
	if strings.HasPrefix(widget.ID, "format $") {
		return []string{"-:" + body}, nil
	}

	if body == "" {
		return []string{"-:"}, nil
	}
	for body != "" {
		buf, s, e := x4extract(ty, body, "")
		body = s
		if e != "" {
			err := &qerror.QError{
				Ref:     []string{"xfile.parse.extract"},
				Version: widget.Release(),
				File:    widget.EditFile(),
				Lineno:  lineno,
				Object:  widget.ID,
				Msg:     []string{e},
			}

			return nil, err
		}
		if len(buf) == 0 {
			continue
		}
		if len(buffer) != 0 && buffer[len(buffer)-1].mode == "I" {
			buf0 := buf[0]
			if buf0.mode == "-" {
				k := strings.Index(buf0.text, "\n")
				if k != -1 {
					buf[0].text = buf0.text[k+1:]
				}
			}
		}
		buffer = append(buffer, buf...)
	}

	ilabels := make([]int, 0)

	for i, x4def := range buffer {
		if x4def.label == "" {
			continue
		}
		ilabels = append(ilabels, i)
	}

	if len(ilabels) == 0 {
		result := packX4(buffer)
		return stringX4(result), nil
	}
	for len(ilabels) != 0 {
		ilabel := ilabels[len(ilabels)-1]
		ilabels = ilabels[:len(ilabels)-1]
		label := buffer[ilabel].label
		ok := false
		for i := ilabel + 1; i < len(buffer); i++ {
			buffer, ok = findLabel(buffer, i, label)
			if ok {
				break
			}
		}
	}

	for i, x4def := range buffer {
		label := x4def.label
		if label == "" {
			continue
		}
		found := false
		for j := i + 1; j < len(buffer); j++ {
			if buffer[j].mode != "L" {
				continue
			}
			if buffer[j].text != label {
				continue
			}
			found = true
			buffer[i].text = strconv.Itoa(j+1) + x4def.text
			break
		}
		if !found {
			err := &qerror.QError{
				Ref:     []string{"xfile.parse.label"},
				Version: widget.Release(),
				File:    widget.EditFile(),
				Lineno:  lineno,
				Object:  widget.ID,
				Msg:     []string{"label `" + label + "` not found"},
			}

			return nil, err
		}
	}
	for i := range buffer {
		if buffer[i].mode != "L" {
			continue
		}
		buffer[i].mode = "-"
		buffer[i].text = ""
	}

	return stringX4(buffer), nil
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

func findLabel(buffer []X4, from int, label string) ([]X4, bool) {
	ok := false
	for i := from; i < len(buffer); i++ {
		x4def := buffer[i]
		mode := x4def.mode
		text := x4def.text
		if mode != "-" && mode != "L" {
			continue
		}
		if mode == "L" && text == label {
			return buffer, true
		}
		if mode == "L" {
			continue
		}

		k := strings.Index(text, label)
		if k == -1 {
			continue
		}
		before := strings.TrimRight(text[:k], " \t\r")
		if before != "" && !strings.HasSuffix(before, "\n") {
			continue
		}
		after := text[k+len(label):]
		if after != "" && strings.TrimLeft(after, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890_") != after {
			continue
		}
		ok = true
		xi := X4{
			mode: "-",
			text: before,
		}
		xii := X4{
			mode: "L",
			text: label,
		}
		xiii := X4{
			mode: "-",
			text: after,
		}
		buffer = append(buffer[:i], append([]X4{xi, xii, xiii}, buffer[i+1:]...)...)
		break
	}
	return buffer, ok
}

func x4extract(ty string, body string, hull string) (buffer []X4, rest string, err string) {
	buffer = make([]X4, 0)
	k := strings.Index(body, "x4_")
	switch k {
	case -1:
		if body != "" {
			buffer = append(buffer, X4{
				mode: "-",
				text: body,
			})
		}
		rest = ""
		return
	case 0:
	default:
		buffer = append(buffer, X4{
			mode: "-",
			text: body[:k],
		})
		body = body[k:]
	}
	body = body[3:]
	rest = strings.TrimLeft(body, "abcdefghijklmnopqrstuvwxyz")
	if len(rest) == len(body) {
		buffer = append(buffer, X4{
			mode: "-",
			text: "x4_",
		})
		return
	}

	verb := body[:len(body)-len(rest)]
	x4inrest := strings.Contains(rest, "x4_")
	args, until, err := qutil.BuildArgs(rest)
	rest = rest[len(until):]
	if err != "" {
		return
	}

	var result X4
	switch verb {
	case "varruntime":
		result, err = x4varruntime(args, ty, hull, x4inrest)
	case "varcoderuntime":
		result, err = x4varcoderuntime(args, ty, hull, x4inrest)
	case "parconstant":
		result, err = x4parconstant(args, ty, hull, x4inrest)
	case "parcode":
		result, err = x4parcode(args, ty, hull, x4inrest)
	case "varcode":
		result, err = x4varcode(args, ty, hull, x4inrest)
	case "vararray":
		result, err = x4vararray(args, ty, hull, x4inrest)
	case "format":
		result, err = x4format(args, ty, hull, x4inrest)
	case "exec":
		result, err = x4exec(args, ty, hull, x4inrest)
	case "if":
		result, err = x4if(args, ty, hull, x4inrest)
	case "select":
		result, err = x4select(args, ty, hull, x4inrest)
	case "lookupinitscreen":
		result, err = x4lookupinitscreen(args, ty, hull, x4inrest)
	case "lookupinitformat":
		result, err = x4lookupinitformat(args, ty, hull, x4inrest)
	}

	if err != "" {
		return
	}
	buffer = append(buffer, result)
	return
}

func x4args(ty string, args []string, hull string) (argums []string, err string) {
	for _, arg := range args {
		if !strings.Contains(arg, "x4_") {
			argums = append(argums, arg)
			continue
		}
		buffer := make([]X4, 0)

		for {
			buffer, arg, err = x4extract(ty, arg, hull)
			if arg == "" || err != "" {
				break
			}
		}
		if err != "" {
			return
		}
		arg := ""
		for _, a := range buffer {
			arg += a.text
		}
		argums = append(argums, arg)
	}
	return
}

func x4varruntime(args []string, ty string, hull string, x4in bool) (result X4, err string) {
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

	if x4in {
		args, err = x4args(ty, args, "varruntime")
		if err != "" {
			return
		}
	}

	enc := ""
	if len(args) == 2 {
		switch args[1] {
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

func x4varcoderuntime(args []string, ty string, hull string, x4in bool) (result X4, err string) {
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
	if x4in {
		args, err = x4args(ty, args, "varcoderuntime")
		if err != "" {
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
		result.text = fmt.Sprintf(`w $$RuntimeC^uwwwscr(%s,"")`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$RuntimeC^uwwwscr(%s,"")`, dnull)
	}
	return
}

func x4varcode(args []string, ty string, hull string, x4in bool) (result X4, err string) {
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

	if x4in {
		args, err = x4args(ty, args, "varcode")
		if err != "" {
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
		result.text = fmt.Sprintf(`w $$TrlCode^uwwwscr(%s)`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$TrlCode^uwwwscr(%s,"")`, dnull)
	}

	return
}

func x4format(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_format cannot be used in other x4 constructions")
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

	if x4in {
		args, err = x4args(ty, args, "format")
		if err != "" {
			return
		}
	}

	// argument transformation
	dnull := ""
	if len(args) != 0 {
		dnull = args[0]
	}
	if dnull == "" {
		err = fmt.Sprintf("x4_format does not work with empty argument")
		return
	}

	// result construction
	result.mode = "X"
	result.text = fmt.Sprintf(`d %%OpenArr^uwwwscr("",""),%%InitRow^uwwwscr("%s",""),%%AddPar^uwwwscr(""),%%ClosArr^uwwwscr d Format^uwwwscr("")`, dnull)

	return
}

func x4parconstant(args []string, ty string, hull string, x4in bool) (result X4, err string) {
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
	if x4in {
		args, err = x4args(ty, args, "parconstant")
		if err != "" {
			return
		}
	}

	switch hull {
	case "", "exec", "if", "select", "lookupinitscreen", "lookupinitformat":
	default:
		err = fmt.Sprintf("x4_parconstant cannot work inside `%s`", hull)
		return
	}

	if x4in {
		args, err = x4args(ty, args, "varcoderuntime")
	}

	enc := ""
	if len(args) == 2 {
		switch args[1] {
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

func x4parcode(args []string, ty string, hull string, x4in bool) (result X4, err string) {
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
	if x4in {
		args, err = x4args(ty, args, "parcode")
		if err != "" {
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
		result.text = fmt.Sprintf(`w $$GetParVl^uwwwscr(PRvar,%s,"CODE")`, dnull)
	} else {
		result.text = fmt.Sprintf(`$$GetParVl^uwwwscr(PRvar,%s,"CODE")`, dnull)
	}
	return
}

func x4vararray(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// check on arguments
	if len(args) > 1 {
		err = fmt.Sprintf("x4_vararray cannot work with %d arguments", len(args))
		return
	}

	if x4in {
		args, err = x4args(ty, args, "vararray")
		if err != "" {
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
		result.text = fmt.Sprintf(`d Format^uwwwscr(%s)`, dnull)
	}
	return
}
func x4exec(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_exec cannot be used in other x4 constructions")
		return
	}
	// check on arguments
	if len(args) != 1 {
		err = fmt.Sprintf("x4_exec works with exactly 1 parameter")
		return
	}
	result.mode = "X"
	result.text = args[0]
	return
}

func x4if(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_if cannot be used in other x4 constructions")
		return
	}

	if len(args) != 2 {
		err = fmt.Sprintf("x4_if works with exactly 2 parameters")
		return
	}
	label := args[0]
	if label == "" {
		err = fmt.Sprintf("x4_if should have a non-emyty label")
	}
	result.mode = "I"
	result.text = ":" + args[1]
	result.label = label

	return
}
func x4select(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_select cannot be used in other x4 constructions")
		return
	}
	if len(args) != 3 {
		err = fmt.Sprintf("x4_select works with exactly 3 parameters")
		return
	}
	dnull := `"` + strings.ReplaceAll(args[0], `"`, `""`) + `"`
	done := `"` + strings.ReplaceAll(args[1], `"`, `""`) + `"`
	result.mode = "X"
	result.text = fmt.Sprintf(`w $s(%s:%s,1:%s)`, args[2], dnull, done)
	return
}
func x4lookupinitscreen(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_lookupinitscreen cannot be used in other x4 constructions")
		return
	}
	if len(args) < 2 {
		err = fmt.Sprintf("x4_lookupinitscreen works with at least 2 parameters")
		return
	}
	if len(args) > 5 {
		err = fmt.Sprintf("x4_lookupinitscreen works with no more than 5 parameters")
		return
	}
	v0 := args[0]
	v1 := args[1]
	v2 := ""
	v3 := ""
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
	v2 = `"` + strings.ReplaceAll(v2, `"`, `""`) + `"`
	v4 = `"` + strings.ReplaceAll(v4, `"`, `""`) + `"`
	result.mode = "X"
	result.text = fmt.Sprintf(`d %%Entry^uluwake(%s,%s,%s,$$Runtime^uwwwscr(%s),%s)`, v0, v1, v2, v3, v4)
	return
}
func x4lookupinitformat(args []string, ty string, hull string, x4in bool) (result X4, err string) {
	// not write
	if hull != "" {
		err = fmt.Sprintf("x4_lookupinitformat cannot be used in other x4 constructions")
		return
	}
	if ty == "screen" {
		err = "x4_lookupinitformat does not work in screens"
	}
	if len(args) < 2 {
		err = fmt.Sprintf("x4_lookupinitformat works with at least 2 parameters")
		return
	}
	if len(args) > 5 {
		err = fmt.Sprintf("x4_lookupinitformat works with no more than 5 parameters")
		return
	}
	v0 := args[0]
	v1 := args[1]
	v2 := ""
	v3 := ""
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
func (widget *Widget) Mumps(batchid string) (mumps qmumps.MUMPS) {

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

	parts := strings.SplitN(widget.Type(), " ", 2)

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

	lines, _ := widget.Resolve()

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
