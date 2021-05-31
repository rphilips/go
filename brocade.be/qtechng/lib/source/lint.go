package source

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qpython "brocade.be/base/python"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// FetchList gets a number of paths
func LintList(version string, paths []string, warnings bool) (infos [][]byte, metas []*qmeta.Meta, errs error) {

	if len(paths) == 0 {
		return
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return nil, nil, err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"source.lintlist.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return nil, nil, err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"source.lintlist.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return nil, nil, err
	}

	type lintdata struct {
		content []byte
		pmeta   *qmeta.Meta
	}

	fn := func(n int) (interface{}, error) {
		p := paths[n]
		source, err := Source{}.New(version, p, true)
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.lintlist.path.nosource"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Path `" + p + "` does not exists"},
			}
			return nil, err
		}
		content, err := source.Lint(warnings)
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.lintlist.path.noread"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Path `" + p + "` unreadable"},
			}
			return nil, err
		}

		pmeta, err := qmeta.Meta{}.New(version, p)
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.lintlist.path.nometa"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Path `" + p + "` not retrievable"},
			}
			return nil, err
		}
		return lintdata{[]byte(content), pmeta}, nil
	}

	result, errorlist := qparallel.NMap(len(paths), -1, fn)
	infos = make([][]byte, len(result))
	metas = make([]*qmeta.Meta, len(result))

	for i, res := range result {
		if errorlist[i] != nil {
			infos[i] = []byte("")
			metas[i] = nil
		} else {
			fres := res.(lintdata)
			infos[i] = fres.content
			metas[i] = fres.pmeta
		}
	}

	errslice := qerror.NewErrorSlice()

	for _, e := range errorlist {
		if e == nil {
			continue
		}
		errslice = append(errslice, e)
	}

	if len(errslice) != 0 {
		errs = errslice
		return
	}

	return
}

// Fetch haalt de data op
func (source *Source) Lint(warnings bool) (info string, err error) {
	natures := source.Natures()
	nolint := natures["nolint"]
	if nolint {
		return "NOLINT", nil
	}
	body, err := source.Fetch()

	if natures["text"] {
		_, _, e := qutil.NoUTF8(bytes.NewReader(body))
		if e != nil {
			info = fmt.Sprintf("`%s` contains non UTF-8 charcter", source.String())
			return
		}
	}

	if err != nil {
		return "", err
	}
	buffer := new(bytes.Buffer)

	switch {
	case natures["lfile"]:
		err = source.Resolve("r", nil, nil, buffer)
		if err != nil {
			return "", err
		}
		return source.LintL(buffer, warnings)
	case natures["dfile"]:
		err = source.Resolve("rl", nil, nil, buffer)
		if err != nil {
			return "", err
		}
		return source.LintD(buffer, warnings)
	case natures["ifile"]:
		err = source.Resolve("rlm", nil, nil, buffer)
		if err != nil {
			return "", err
		}
		return source.LintI(buffer, warnings)
	case natures["bfile"]:
		err = source.Resolve("rilm", nil, nil, buffer)
		if err != nil {
			return "", err
		}
		return source.LintB(buffer, warnings)
	}

	err = source.Resolve("rilm", nil, nil, buffer)
	if err != nil {
		return "", err
	}
	ext := path.Ext(source.String())
	switch ext {
	case ".py":
		return source.LintPy(buffer, warnings)
	case ".x":
		return source.LintX(buffer, warnings)
	case ".m":
		return source.LintM(buffer, warnings)
	}

	return "", nil

}

//Parse Python File
func (source *Source) LintPy(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	body := buffer.Bytes()
	release := source.Release()
	fs, pyscript := release.SourcePlace(source.String())
	pyscript, _ = fs.RealPath(pyscript)
	py := qutil.GetPy(pyscript)
	tmppy, _ := qfs.TempFile("", filepath.Base(pyscript)+"_")

	tmppy += ".py"
	qfs.Store(tmppy, body, "")
	e := qpython.Compile(tmppy, py == "py3")
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	return info, nil
}

// Parse BFile
func (source *Source) LintB(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	bfile := new(qofile.BFile)
	bfile.Source = source.String()
	bfile.Version = source.Release().String()
	preamble, objs, e := bfile.Parse(buffer.Bytes(), true)
	info = "OK"

	if e != nil {
		info = e.Error()
	}
	if info == "" {
		info = handleObjects(objs)
	}

	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}

	return info, nil
}

// Parse DFile
func (source *Source) LintD(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	dfile := new(qofile.DFile)
	dfile.Source = source.String()
	dfile.Version = source.Release().String()
	preamble, objs, e := dfile.Parse(buffer.Bytes(), true)
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	if info == "" {
		info = handleObjects(objs)
	}
	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil
}

// Parse IFile
func (source *Source) LintI(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	ifile := new(qofile.IFile)
	ifile.Source = source.String()
	ifile.Version = source.Release().String()
	preamble, objs, e := ifile.Parse(buffer.Bytes(), true)
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	if info == "" {
		info = handleObjects(objs)
	}
	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil
}

// Parse LFile
func (source *Source) LintL(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	lfile := new(qofile.LFile)
	lfile.Source = source.String()
	lfile.Version = source.Release().String()
	preamble, objs, e := lfile.Parse(buffer.Bytes(), true)
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	if info == "" {
		info = handleObjects(objs)
	}
	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil

}

// Parse MFile
func (source *Source) LintM(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	preamble := ""
	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return
}

// Parse XFile
func (source *Source) LintX(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	xfile := new(qofile.BFile)
	xfile.Source = source.String()
	xfile.Version = source.Release().String()
	preamble, _, e := xfile.Parse(buffer.Bytes(), true)
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	if info == "" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil
}

func handleObjects(objs []qobject.Object) string {
	if len(objs) == 0 {
		return ""
	}
	m := make(map[string]bool)
	version := ""
	for _, obj := range objs {
		name := obj.String()
		if m[name] {
			return fmt.Sprintf("Object `%s` occurs more than one in `%s`", name, obj.EditFile())
		}
		err := obj.Lint()
		if err != nil && len(err) > 0 {
			return fmt.Sprintf("Object `%s` occurs more than one in `%s`", name, err[0].Error())
		}
		if version == "" {
			version = obj.Release()
		}
	}
	if version != "" {
		return "No legal version found"
	}

	fn := func(n int) (interface{}, error) {
		object := objs[n]
		editfile := object.EditFile()
		qobject.Fetch(object)
		editfile2 := object.EditFile()
		if editfile2 != "" && editfile2 != editfile {
			return fmt.Sprintf("Object `%s` is also defined in `%s`", object.String(), editfile2), nil
		}
		return "", nil
	}
	resultlist, _ := qparallel.NMap(len(objs), 1, fn)

	infos := make([]string, 0)
	for _, r := range resultlist {
		s := r.(string)
		if s == "" {
			continue
		}
		infos = append(infos, s)
	}

	return strings.Join(infos, "; ")
}
