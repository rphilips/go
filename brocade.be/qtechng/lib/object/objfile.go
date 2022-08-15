package object

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// OFile stelt een bestand met objecten voor
type OFile interface {
	String() string
	Type() string
	Release() string
	SetRelease(release string)
	EditFile() string
	SetEditFile(editfile string)
	Parse(blob []byte, decomment bool) (string, []Object, error)
	SetComment(string)
	Comment() string
	Objects() []Object
	SetObjects([]Object)
	Sort()
}

// Loads an objectfile
func Loads(ofile OFile, blob []byte, decomment bool) (err error) {
	fname := ofile.EditFile()
	if blob == nil {
		blob, err = os.ReadFile(fname)
		if err != nil {
			e := &qerror.QError{
				Ref:    []string{"objfile.loads.read"},
				QPath:  fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    qerror.ErrorMsg(err),
			}
			return e
		}
	}

	blob = qutil.About(blob)

	preamble, objects, e := ofile.Parse(blob, decomment)
	if e != nil {
		msg, lineno := qerror.ExtractEMsg(e, fname, blob)

		err := &qerror.QError{
			Ref:    []string{"objfile.loads.parse"},
			QPath:  fname,
			Lineno: lineno,
			Type:   "Error",
			Msg:    msg,
		}
		return err
	}

	// check on doubles

	found := make(map[string]int)

	for nr, obj := range objects {
		id := obj.Name()
		if found[id] != 0 {
			lineno, _ := strconv.Atoi(obj.Lineno())
			err = &qerror.QError{
				Ref:    []string{"objfile.lint.double"},
				File:   fname,
				Lineno: lineno,
				Object: obj.String(),
				Type:   "Error",
				Msg:    []string{"`" + id + "` found at " + objects[found[id]-1].Lineno() + " and " + obj.Lineno()},
			}
			return
		}
		found[id] = nr + 1
	}

	ofile.SetComment(preamble)
	release := ofile.Release()
	editfile := ofile.EditFile()
	for _, obj := range objects {
		if release != "" && obj.Release() == "" {
			obj.SetRelease(release)
		}
		if editfile != "" && obj.EditFile() == "" {
			obj.SetEditFile(editfile)
		}
	}
	ofile.SetObjects(objects)

	return nil
}

// Format formateer een bestand
func Format(ofile OFile) (lines []string) {
	ofile.Sort()
	lines = []string{ofile.Comment()}

	for _, obj := range ofile.Objects() {
		lines = append(lines, "", obj.Format())
	}
	return
}

//StoreFileObjects stores a list of object form an OFile
func StoreFileObjects(ofile OFile) (changedmap map[string]bool, errorlist []error) {
	objs := ofile.Objects()
	return StoreList(objs)
}

// Lint parst een bestand zonder naar het repository te gaan
func Lint(ofile OFile, blob []byte, current []byte) (err error) {
	fname := ofile.EditFile()
	if len(blob) == 0 {
		blob, err = os.ReadFile(fname)
		if err != nil {
			e := &qerror.QError{
				Ref:    []string{"objfile.lint.read"},
				QPath:  fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err.Error()},
			}
			return e
		}
	}
	blob = qutil.About(blob)
	// check on UTF-8
	body, badutf8, e := qutil.NoUTF8(bytes.NewReader(blob))
	if e != nil {
		err = &qerror.QError{
			Ref:    []string{"objfile.lint"},
			QPath:  fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return
	}
	if len(badutf8) != 0 {
		err = &qerror.QError{
			Ref:    []string{"objfile.lint.utf8"},
			QPath:  fname,
			Lineno: badutf8[0][0],
			Type:   "Error",
			Msg:    []string{"Contains non-UTF8"},
		}
		return
	}

	ep := Loads(ofile, body, true)

	if ep != nil {
		err := &qerror.QError{
			Ref:  []string{"objfile.lint.loads"},
			File: fname,
			Type: "Error",
			Msg:  []string{},
		}
		return qerror.QErrorTune(ep, err)
	}

	// preamble

	preamble := ofile.Comment()

	if strings.TrimSpace(strings.Trim(preamble, "/")) == "" {
		err = &qerror.QError{
			Ref:    []string{"objfile.lint.preamble"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{"Preamble is empty"},
		}
		return
	}

	// check on doubles

	found := make(map[string]int)

	objs := ofile.Objects()

	for nr, obj := range objs {
		id := obj.Name()
		if found[id] != 0 {
			lineno, _ := strconv.Atoi(obj.Lineno())
			err = &qerror.QError{
				Ref:    []string{"objfile.lint.double"},
				File:   fname,
				Lineno: lineno,
				Object: obj.String(),
				Type:   "Error",
				Msg:    []string{"`" + id + "` found at " + objs[found[id]-1].Lineno() + " and " + obj.Lineno()},
			}
			return
		}
		found[id] = nr + 1
	}

	// /individual tests
	errslice := qerror.NewErrorSlice()

	for _, obj := range objs {
		errs := obj.Lint()
		for _, e := range errs {
			if e != nil {
				errslice = append(errslice, e)
			}
		}
	}

	// local
	release := ofile.Release()
	if release == "" || fname == "" || len(current) == 0 {
		if len(errslice) == 0 {
			return nil
		}
		return errslice
	}

	// in repository
	tocheck := make(map[string]Object)

	for _, obj := range objs {
		id := obj.String()
		tocheck[id] = obj
	}

	_, obs, ep := ofile.Parse(current, true)

	if ep == nil {
		for _, obj := range obs {
			id := obj.String()
			delete(tocheck, id)
		}
	}

	if len(tocheck) != 0 {
		objlist := make([]Object, 0)
		for _, obj := range tocheck {
			obj.SetRelease(release)
			objlist = append(objlist, obj)
		}
		sourcemap := SourceList(objlist)
		for name, editfile := range sourcemap {
			if editfile == "" || editfile == fname {
				continue
			}
			obj := tocheck[name]
			lineno, _ := strconv.Atoi(obj.Lineno())
			e := &qerror.QError{
				Ref:    []string{"objfile.lint" + ".otherdefinition"},
				File:   fname,
				Lineno: lineno,
				Object: obj.String(),
				Type:   "Error",
				Msg:    []string{"Already defined in `" + editfile + "`"},
			}
			errslice = append(errslice, e)
		}
	}

	return errslice
}

// LintObject parst objecten zonder naar het repository te gaan
func LintObjects(ofile OFile) (err error) {
	// check on doubles
	fname := ofile.EditFile()
	found := make(map[string]int)

	objs := ofile.Objects()

	for nr, obj := range objs {
		id := obj.Name()
		if found[id] != 0 {
			lineno, _ := strconv.Atoi(obj.Lineno())
			err = &qerror.QError{
				Ref:    []string{"objfile.lint.double"},
				File:   fname,
				Lineno: lineno,
				Object: obj.String(),
				Type:   "Error",
				Msg:    []string{"`" + id + "` found at " + objs[found[id]-1].Lineno() + " and " + obj.Lineno()},
			}
			return
		}
		found[id] = nr + 1
	}

	// /individual tests
	errslice := qerror.NewErrorSlice()

	for _, obj := range objs {
		errs := obj.Lint()
		for _, e := range errs {
			if e != nil {
				errslice = append(errslice, e)
			}
		}
	}
	if len(errslice) == 0 {
		return nil
	}
	return errslice
}
