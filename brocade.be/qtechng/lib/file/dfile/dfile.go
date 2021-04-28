package dfile

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"

	qregistry "brocade.be/base/registry"
)

var pardoc = make(map[string]string)

// DFile stelt een bestand met macro's voor (*.d)
type DFile struct {
	Preamble string   // Commentaar in het begin van de file
	Macros   []*Macro // Lijst met macro's
	Source   string   `json:"source"`  // Editfile
	Version  string   `json:"release"` // Version
}

// Type interface method
func (df *DFile) Type() string {
	return "dfile"
}

// Release interface method
func (df *DFile) Release() string {
	return df.Version
}

// SetRelease interface method
func (df *DFile) SetRelease(release string) {
	df.Version = release
}

// EditFile interface method
func (df *DFile) EditFile() string {
	return df.Source
}

// SetEditFile of macro
func (df *DFile) SetEditFile(source string) {
	df.Source = source
}

// Comment interface method
func (df *DFile) Comment() string {
	return df.Preamble
}

// SetComment of macro file
func (df *DFile) SetComment(preamble string) {
	df.Preamble = preamble
}

// String interface method
func (df *DFile) String() string {
	return df.Source
}

// Parse parst een []byte
func (df *DFile) Parse(blob []byte, decomment bool) (preamble string, objs []qobject.Object, err error) {
	fname := df.EditFile()
	x, err := Parse(fname, blob)
	if err != nil {
		return
	}
	y := x.(DFile)
	preamble = y.Preamble
	if len(y.Macros) == 0 {
		return
	}
	if len(pardoc) == 0 && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'W') {
		pardoc["_"] = ""
		x := filepath.Join(strings.SplitN(qregistry.Registry["qtechng-support-project"], "/", -1)[1:]...)
		name := filepath.Join(qregistry.Registry["qtechng-work-dir"], x, "pardoc.json")
		data, e := os.ReadFile(name)
		if e == nil {
			json.Unmarshal(data, &pardoc)
		}
	}
	cmt := make(map[string]string)
	for k, v := range pardoc {
		cmt["$"+k] = v
	}

	objs = make([]qobject.Object, len(y.Macros))
	release := df.Release()
	for k, macro := range y.Macros {
		for i, param := range macro.Params {
			doc := strings.TrimSpace(param.Doc)
			pid := param.ID
			if doc != "" && strings.Trim(pid, "$1234567890") != "" {
				cmt[pid] = doc
			}
			if doc != "" {
				continue
			}
			sdoc := cmt[pid]
			param.Doc = sdoc
			macro.Params[i] = param
		}
		macro.SetRelease(release)
		macro.SetEditFile(fname)
		objs[k] = macro
	}
	return
}

// Objects interface method
func (df *DFile) Objects() []qobject.Object {
	if df.Macros == nil {
		return nil
	}
	objs := make([]qobject.Object, len(df.Macros))

	for i, macro := range df.Macros {
		objs[i] = macro
	}
	return objs
}

// SetObjects interface method
func (df *DFile) SetObjects(objects []qobject.Object) {
	if len(objects) == 0 {
		df.Macros = nil
		return
	}
	fname := df.EditFile()
	release := df.Release()
	macros := make([]*Macro, len(objects))
	for i, obj := range objects {
		obj.SetRelease(release)
		obj.SetEditFile(fname)
		macros[i] = obj.(*Macro)
	}
	df.Macros = macros
}

// Sort interface method
func (df *DFile) Sort() {
	objs := df.Macros
	macros := make([]*Macro, 0)
	found := make(map[string]bool)
	for i, obj := range objs {
		name := obj.Name()
		if found[name] {
			continue
		}
		if len(name) < 4 {
			macros = append(macros, obj)
			found[name] = true
			continue
		}
		k := strings.IndexAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		prefix := ""
		if k != -1 {
			prefix = name[:k]
		}
		if prefix != "is" && prefix != "get" && prefix != "gen" && prefix != "set" && prefix != "del" && prefix != "upd" {
			macros = append(macros, obj)
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
				macros = append(macros, oj)
				found[name] = true
			}
		}
	}
	df.Macros = macros
}

// Format formats a B file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"dfile.format.read"},
				File:   fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err1.Error()},
			}
			return e
		}
	}

	objfile := new(DFile)
	objfile.SetEditFile(fname)
	err := qobject.Loads(objfile, blob, true)
	if err != nil {
		output.Write(blob)
		return nil
	}

	blob = qutil.About(blob)
	// check on UTF-8
	body, badutf8, e := qutil.NoUTF8(bytes.NewReader(blob))
	if e != nil {
		err := &qerror.QError{
			Ref:    []string{"dfile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"dfile.format.utf8"},
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
	delims := []string{
		"macro",
	}

	macros := qutil.BlobSplit(body, delims, false)
	wrote := false

	if len(macros) != 0 {
		m0 := bytes.TrimSpace(macros[0])
		output.Write(m0)
	}

	if len(pardoc) != 0 && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'W') {
		x := filepath.Join(strings.SplitN(qregistry.Registry["qtechng-support-project"], "/", -1)[1:]...)
		name := filepath.Join(qregistry.Registry["qtechng-work-dir"], x, "pardoc.json")
		data, e := os.ReadFile(name)
		if e == nil {
			json.Unmarshal(data, &pardoc)
		}
		pardoc["_"] = ""
	}
	cmt := make(map[string]string)
	for k, v := range pardoc {
		cmt["$"+k] = v
	}

	first := true

	for _, part := range macros {
		if first {
			first = false
			continue
		}
		macro := new(Macro)
		err := macro.Loads(part)
		output.WriteString("\n\n")
		wrote = true
		if err != nil {
			m0 := bytes.TrimSpace(part)
			output.Write(m0)
			continue
		}
		for i, param := range macro.Params {
			doc := strings.TrimSpace(param.Doc)
			pid := param.ID
			if doc != "" && strings.Trim(pid, "$1234567890") != "" {
				cmt[pid] = doc
			}
			if doc != "" {
				continue
			}
			sdoc := cmt[pid]
			param.Doc = sdoc
			macro.Params[i] = param
		}
		output.WriteString(macro.Format())
	}

	if wrote {
		output.WriteString("\n")
	}
	return nil
}
