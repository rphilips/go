package xfile

import (
	"bytes"
	"os"
	"regexp"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

var (
	retagb = regexp.MustCompile(`^\s*(format|screen|text)\s+([^:\s]+)\s*:(.*)$`)
	ws     = regexp.MustCompile(`\s+`)
)

// XFile stelt een bestand met screens en formats voor (*.x)
type XFile struct {
	Preamble string    // Commentaar in het begin van de file
	Widgets  []*Widget // Lijst met Widgets
	Source   string    `json:"source"`  // Editfile
	Version  string    `json:"release"` // Version
}

// Type interface method
func (xf *XFile) Type() string {
	return "XFile"
}

// Release interface method
func (xf *XFile) Release() string {
	return xf.Version
}

// SetRelease interface method
func (xf *XFile) SetRelease(release string) {
	xf.Version = release
}

// EditFile interface method
func (xf *XFile) EditFile() string {
	return xf.Source
}

// SetEditFile of Widget
func (xf *XFile) SetEditFile(source string) {
	xf.Source = source
}

// Comment interface method
func (xf *XFile) Comment() string {
	return xf.Preamble
}

// SetComment of Widget file
func (xf *XFile) SetComment(preamble string) {
	xf.Preamble = preamble
}

// String interface method
func (xf *XFile) String() string {
	return xf.Source
}

// Parse parst een []byte
func (xf *XFile) Parse(blob []byte) (preamble string, objs []qobject.Object, err error) {
	fname := xf.EditFile()
	x, err := Parse(fname, blob)
	if err != nil {
		return
	}
	y := x.(XFile)
	preamble = y.Preamble
	if len(y.Widgets) == 0 {
		return
	}
	objs = make([]qobject.Object, len(y.Widgets))
	release := xf.Release()
	for k, widget := range y.Widgets {
		widget.SetRelease(release)
		widget.SetEditFile(fname)
		objs[k] = widget
	}
	return
}

// Objects interface method
func (xf *XFile) Objects() []qobject.Object {
	if xf.Widgets == nil {
		return nil
	}
	objs := make([]qobject.Object, len(xf.Widgets))

	for i, Widget := range xf.Widgets {
		objs[i] = Widget
	}
	return objs
}

// SetObjects interface method
func (xf *XFile) SetObjects(objects []qobject.Object) {
	if len(objects) == 0 {
		xf.Widgets = nil
		return
	}
	fname := xf.EditFile()
	release := xf.Release()
	Widgets := make([]*Widget, len(objects))
	for i, obj := range objects {
		obj.SetRelease(release)
		obj.SetEditFile(fname)
		Widgets[i] = obj.(*Widget)
	}
	xf.Widgets = Widgets
}

// Sort interface method
func (xf *XFile) Sort() {
	objs := xf.Widgets
	Widgets := make([]*Widget, 0)
	found := make(map[string]bool)
	for i, obj := range objs {
		name := obj.Name()
		if found[name] {
			continue
		}
		if len(name) < 4 {
			Widgets = append(Widgets, obj)
			found[name] = true
			continue
		}
		k := strings.IndexAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		prefix := ""
		if k != -1 {
			prefix = name[:k]
		}
		if prefix != "is" && prefix != "get" && prefix != "gen" && prefix != "set" && prefix != "del" && prefix != "upd" {
			Widgets = append(Widgets, obj)
			found[name] = true
			continue
		}
		needle := name[k:]
		for _, prefix := range []string{"is", "get", "gen", "set", "upd", "del"} {
			search := prefix + needle
			if found[search] {
				continue
			}
			for _, oj := range objs[i:] {
				name := oj.Name()
				if name != search {
					continue
				}
				Widgets = append(Widgets, oj)
				found[name] = true
			}
		}
	}
	xf.Widgets = Widgets
}

// Format formats a X file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"xfile.format.read"},
				File:   fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err1.Error()},
			}
			return e
		}
	}
	blob = qutil.About(blob)
	// check on UTF-8
	body, badutf8, e := qutil.NoUTF8(bytes.NewReader(blob))
	if e != nil {
		err := &qerror.QError{
			Ref:    []string{"xfile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"xfile.format.utf8"},
			File:   fname,
			Lineno: badutf8[0][0],
			Type:   "Error",
			Msg:    []string{"Contains non-UTF8"},
		}
		return err
	}
	if len(body) == 0 {
		return nil
	}

	screfors := qutil.BlobSplit(body, []string{"screen", "text", "format"}, false)
	wrote := false
	if len(screfors) != 0 {
		output.Write(bytes.TrimSpace(screfors[0]))
		wrote = true
	}

	for _, mode := range [3]string{"text", "screen", "format"} {
		bmode := []byte(mode)
		for _, screfor := range screfors {
			screfor = bytes.TrimSpace(screfor)
			if !bytes.HasPrefix(screfor, bmode) {
				continue
			}

			widget := new(Widget)
			err := widget.Loads(screfor)
			output.WriteString("\n\n")
			if err != nil {
				output.Write(bytes.TrimSpace(screfor))
			} else {
				output.WriteString(widget.ID + ":\n")
				output.WriteString(widget.Body)
			}
			wrote = true
		}
	}

	if wrote {
		output.WriteString("\n")
	}
	return nil
}
