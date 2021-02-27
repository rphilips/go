package ifile

import (
	"bytes"
	"os"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// IFile stelt een bestand met lgcodes voor (*.l)
type IFile struct {
	Preamble string     // Commentaar in het begin van de file
	Includes []*Include // Lijst met lgcodes
	Source   string     `json:"source"`  // Editfile
	Version  string     `json:"release"` // Version
}

// Type interface method
func (df *IFile) Type() string {
	return "ifile"
}

// Release interface method
func (df *IFile) Release() string {
	return df.Version
}

// SetRelease interface method
func (df *IFile) SetRelease(release string) {
	df.Version = release
}

// EditFile interface method
func (df *IFile) EditFile() string {
	return df.Source
}

// SetEditFile of include
func (df *IFile) SetEditFile(source string) {
	df.Source = source
}

// Comment interface method
func (df *IFile) Comment() string {
	return df.Preamble
}

// SetComment of include file
func (df *IFile) SetComment(preamble string) {
	df.Preamble = preamble
}

// String interface method
func (df *IFile) String() string {
	return df.Source
}

// Parse parst een []byte
func (df *IFile) Parse(blob []byte) (preamble string, objs []qobject.Object, err error) {
	fname := df.EditFile()

	x, err := Parse(fname, blob)
	if err != nil {
		return
	}
	y := x.(IFile)
	preamble = y.Preamble
	if len(y.Includes) == 0 {
		return
	}

	objs = make([]qobject.Object, len(y.Includes))
	release := df.Release()
	for k, include := range y.Includes {
		include.SetRelease(release)
		include.SetEditFile(fname)
		objs[k] = include
	}
	return
}

// Objects interface method
func (df *IFile) Objects() []qobject.Object {
	if df.Includes == nil {
		return nil
	}
	objs := make([]qobject.Object, len(df.Includes))

	for i, include := range df.Includes {
		objs[i] = include
	}
	return objs
}

// SetObjects interface method
func (df *IFile) SetObjects(objects []qobject.Object) {
	if len(objects) == 0 {
		df.Includes = nil
		return
	}
	fname := df.EditFile()
	release := df.Release()
	includes := make([]*Include, len(objects))
	for i, obj := range objects {
		obj.SetRelease(release)
		obj.SetEditFile(fname)
		includes[i] = obj.(*Include)
	}
	df.Includes = includes
}

// Sort interface method
func (df *IFile) Sort() {
}

// Format formats a B file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"ifile.format.read"},
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
			Ref:    []string{"ifile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"ifile.format.utf8"},
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
		"include",
	}

	includes := qutil.BlobSplit(body, delims, false)
	wrote := false
	if len(includes) != 0 {
		output.Write(bytes.TrimSpace(includes[0]))
		wrote = true
	}
	first := true
	for _, part := range includes {
		if first {
			first = false
			continue
		}
		include := new(Include)
		err := include.Loads(part)
		output.WriteString("\n\n")
		if err != nil {
			output.Write(bytes.TrimSpace(part))
		} else {
			output.WriteString(include.Format())
		}
		wrote = true
	}

	if wrote {
		output.WriteString("\n")
	}
	return nil
}
