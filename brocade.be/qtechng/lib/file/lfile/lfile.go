package lfile

import (
	"bytes"
	"os"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// LFile stelt een bestand met lgcodes voor (*.l)
type LFile struct {
	Preamble string    // Commentaar in het begin van de file
	Lgcodes  []*Lgcode // Lijst met lgcodes
	Source   string    `json:"source"`  // Editfile
	Version  string    `json:"release"` // Version
}

// Type interface method
func (lf *LFile) Type() string {
	return "lfile"
}

// Release interface method
func (lf *LFile) Release() string {
	return lf.Version
}

// SetRelease interface method
func (lf *LFile) SetRelease(release string) {
	lf.Version = release
}

// EditFile interface method
func (lf *LFile) EditFile() string {
	return lf.Source
}

// SetEditFile of macro
func (lf *LFile) SetEditFile(source string) {
	lf.Source = source
}

// Comment interface method
func (lf *LFile) Comment() string {
	return lf.Preamble
}

// SetComment of macro file
func (lf *LFile) SetComment(preamble string) {
	lf.Preamble = preamble
}

// String interface method
func (lf *LFile) String() string {
	return lf.Source
}

// Parse parst een []byte
func (lf *LFile) Parse(blob []byte) (preamble string, objs []qobject.Object, err error) {
	fname := lf.EditFile()
	x, err := Parse(fname, blob)
	if err != nil {
		return
	}
	y := x.(LFile)
	preamble = y.Preamble
	if len(y.Lgcodes) == 0 {
		return
	}

	objs = make([]qobject.Object, len(y.Lgcodes))
	release := lf.Release()
	for k, lgcode := range y.Lgcodes {
		lgcode.SetRelease(release)
		lgcode.SetEditFile(fname)
		objs[k] = lgcode
	}
	return
}

// Objects interface method
func (lf *LFile) Objects() []qobject.Object {
	if lf.Lgcodes == nil {
		return nil
	}
	objs := make([]qobject.Object, len(lf.Lgcodes))

	for i, lgcode := range lf.Lgcodes {
		objs[i] = lgcode
	}
	return objs
}

// SetObjects interface method
func (lf *LFile) SetObjects(objects []qobject.Object) {
	if len(objects) == 0 {
		lf.Lgcodes = nil
		return
	}
	fname := lf.EditFile()
	release := lf.Release()
	lgcodes := make([]*Lgcode, len(objects))
	for i, obj := range objects {
		obj.SetRelease(release)
		obj.SetEditFile(fname)
		lgcodes[i] = obj.(*Lgcode)
	}
	lf.Lgcodes = lgcodes
}

// Sort interface method
func (lf *LFile) Sort() {
	objs := lf.Lgcodes
	lgcodes := make([]*Lgcode, 0)
	found := make(map[string]int)

	// maak een "look-ahead"
	for nr, lgcode := range objs {
		id := lgcode.ID
		found[id] = nr
	}

	done := make(map[string]bool)

	for _, lgcode := range objs {
		id := lgcode.ID
		if done[id] {
			continue
		}
		count := strings.Count(id, ".")

		if count != 0 {
			parts := strings.SplitN(id, ".", 3)
			// behandel de namespace
			nmspace := parts[0] + "."
			if !done[nmspace] {
				done[nmspace] = true
				nr, ok := found[nmspace]
				if ok {
					lgcodes = append(lgcodes, objs[nr])
				}
			}
			// behandel het textfragment
			text := parts[1]
			if text != "" {
				text = nmspace + text
				if !done[text] {
					done[text] = true
					nr, ok := found[text]
					if ok {
						lgcodes = append(lgcodes, objs[nr])
					}
				}
				// ... en scope
				scope := text + ".scope"
				if !done[scope] {
					done[scope] = true
					nr, ok := found[scope]
					if ok {
						lgcodes = append(lgcodes, objs[nr])
					}
				}
			}
		}
		if !done[id] {
			lgcodes = append(lgcodes, lgcode)
			done[id] = true
		}
	}
	lf.Lgcodes = lgcodes
}

// Format formats a L-file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"lfile.format.read"},
				File:   fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err1.Error()},
			}
			return e
		}
	}

	objfile := new(LFile)
	objfile.SetEditFile(fname)
	err := qobject.Loads(objfile, blob)
	if err != nil {
		output.Write(blob)
		return nil
	}

	blob = qutil.About(blob)
	// check on UTF-8
	body, badutf8, e := qutil.NoUTF8(bytes.NewReader(blob))
	if e != nil {
		err := &qerror.QError{
			Ref:    []string{"lfile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"lfile.format.utf8"},
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
		"lgcode",
	}

	lgcodes := qutil.BlobSplit(body, delims, false)
	wrote := false
	if len(lgcodes) != 0 {
		output.Write(bytes.TrimSpace(lgcodes[0]))
		wrote = true
	}
	first := true
	for _, part := range lgcodes {
		if first {
			first = false
			continue
		}
		lgcode := new(Lgcode)
		err := lgcode.Loads(part)
		output.WriteString("\n\n")
		if err != nil {
			output.Write(bytes.TrimSpace(part))
		} else {
			output.WriteString(lgcode.Format())
		}
		wrote = true
	}

	if wrote {
		output.WriteString("\n")
	}
	return nil
}
