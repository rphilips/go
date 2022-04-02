package bfile

import (
	"bytes"
	"os"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// BFile stelt een bestand met brob's voor (*.d)
type BFile struct {
	Preamble string  // Commentaar in het begin van de file
	Brobs    []*Brob // Lijst met Brocade objecten
	Source   string  `json:"source"`  // Editfile
	Version  string  `json:"release"` // Version
}

// Type interface method
func (bf *BFile) Type() string {
	return "bfile"
}

// Release interface method
func (bf *BFile) Release() string {
	return bf.Version
}

// SetRelease interface method
func (bf *BFile) SetRelease(release string) {
	bf.Version = release
}

// EditFile interface method
func (bf *BFile) EditFile() string {
	return bf.Source
}

// SetEditFile of brob
func (bf *BFile) SetEditFile(source string) {
	bf.Source = source
}

// Comment interface method
func (bf *BFile) Comment() string {
	return bf.Preamble
}

// SetComment of brob file
func (bf *BFile) SetComment(preamble string) {
	bf.Preamble = preamble
}

// String interface method
func (bf *BFile) String() string {
	return bf.Source
}

// Parse parst een []byte
func (bf *BFile) Parse(blob []byte, decomment bool) (preamble string, objs []qobject.Object, err error) {
	fname := bf.EditFile()
	var x interface{}
	if decomment {
		x, err = Parse(fname, qutil.Decomment(blob).Bytes())
	} else {
		x, err = Parse(fname, blob)
	}
	if err != nil {
		return
	}
	y := x.(BFile)
	preamble = y.Preamble
	if len(y.Brobs) == 0 {
		return
	}

	release := bf.Release()
	for _, brob := range y.Brobs {
		brob.SetRelease(release)
		brob.SetEditFile(fname)
		objs = append(objs, brob)
	}
	return
}

// Objects interface method
func (bf *BFile) Objects() []qobject.Object {
	if bf.Brobs == nil {
		return nil
	}
	objs := make([]qobject.Object, len(bf.Brobs))

	for i, brob := range bf.Brobs {
		objs[i] = brob
	}
	return objs
}

// SetObjects interface method
func (bf *BFile) SetObjects(objects []qobject.Object) {
	if len(objects) == 0 {
		bf.Brobs = nil
		return
	}
	fname := bf.EditFile()
	release := bf.Release()
	brobs := make([]*Brob, len(objects))
	for i, obj := range objects {
		obj.SetRelease(release)
		obj.SetEditFile(fname)
		brobs[i] = obj.(*Brob)
	}
	bf.Brobs = brobs
}

// Sort interface method
func (bf *BFile) Sort() {

}

// Format formats a B file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"bfile.format.read"},
				File:   fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err1.Error()},
			}
			return e
		}
	}

	objfile := new(BFile)
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
			Ref:    []string{"bfile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"bfile.format.utf8"},
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
		"oaiset",
		"oai",
		"mprocess",
		"mailtrg",
		"usergroup",
		"lookup",
		"history",
		"meta",
		"ujson",
		"listattribute",
		"listidentity",
		"listdownloadtype",
		"cg",
		"loi",
		"search",
		"listsorttype",
		"nodeattribute",
		"listconversion",
	}

	brobs := qutil.BlobSplit(body, delims, false)
	wrote := false
	if len(brobs) != 0 && len(brobs[0]) != 0 {
		output.Write(brobs[0])
		wrote = true
	}
	first := true
	for _, part := range brobs {
		if first {
			first = false
			continue
		}
		brob := new(Brob)
		err := brob.Loads(part)
		if wrote {
			output.WriteString("\n\n")
		}
		if err != nil {
			output.Write(bytes.TrimSpace(part))
		} else {
			output.WriteString(brob.Format())
		}
		wrote = true
	}

	if wrote {
		output.WriteString("\n")
	}
	return nil
}
