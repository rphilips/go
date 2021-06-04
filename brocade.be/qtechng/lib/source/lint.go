package source

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qphp "brocade.be/base/php"
	qpython "brocade.be/base/python"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
	qyaml "gopkg.in/yaml.v2"
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
	if err != nil {
		return "", err
	}

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
	case strings.HasSuffix(source.String(), "/brocade.json"):
		buffer.Write(body)
		return source.LintBrocadeJson(buffer, warnings)
	case natures["lfile"]:
		err = source.Resolve("r", nil, nil, buffer, true)
		if err != nil {
			return "", err
		}
		return source.LintL(buffer, warnings)
	case natures["dfile"]:
		err = source.Resolve("rl", nil, nil, buffer, true)
		if err != nil {
			return "", err
		}
		return source.LintD(buffer, warnings)
	case natures["ifile"]:
		err = source.Resolve("rlm", nil, nil, buffer, true)
		if err != nil {
			return "", err
		}
		return source.LintI(buffer, warnings)
	case natures["bfile"]:
		err = source.Resolve("rilm", nil, nil, buffer, true)
		if err != nil {
			return "", err
		}
		return source.LintB(buffer, warnings)
	}
	if natures["mfile"] {
		err = source.Resolve("rilm", nil, nil, buffer, true)
	} else {
		err = source.Resolve("rilm", nil, nil, buffer, false)
	}
	if err != nil {
		return err.Error(), nil
	}
	ext := path.Ext(source.String())
	switch ext {
	case ".php", ".phtml":
		return source.LintPHP(buffer, warnings)
	case ".yaml", ".yml":
		return source.LintYAML(buffer, warnings)
	case ".json":
		return source.LintJSON(buffer, warnings)
	case ".xml":
		return source.LintXML(buffer, warnings)
	case ".py":
		return source.LintPy(buffer, warnings)
	case ".x":
		return source.LintX(buffer, warnings)
	case ".m":
		return source.LintM(buffer, warnings)
	}

	return "OK", nil

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
		info = e.Error() + " [" + py + "]"
	}
	if info == "OK" {
		qfs.Rmpath(tmppy)
	}
	return info, nil
}

//Parse PHP File
func (source *Source) LintPHP(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	body := buffer.Bytes()
	release := source.Release()
	fs, phpscript := release.SourcePlace(source.String())
	phpscript, _ = fs.RealPath(phpscript)
	tmpphp, _ := qfs.TempFile("", filepath.Base(phpscript)+"_")

	tmpphp += ".php"
	qfs.Store(tmpphp, body, "")
	e := qphp.Compile(tmpphp)
	info = "OK"
	if e != nil {
		info = e.Error()
	}
	if info == "OK" {
		qfs.Rmpath(tmpphp)
	}
	return info, nil
}

// brocade.json

func (source *Source) LintBrocadeJson(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	body := buffer.Bytes()
	config := new(qproject.Config)
	e := json.Unmarshal(body, config)
	if e != nil {
		info = fmt.Sprintf("Not valid JSON in `%s`: %s", source.String(), e.Error())
		return info, nil
	}
	if !qproject.IsValidConfig(body) {
		info = fmt.Sprintf("`%s` is not a valid configuration file", source.String())
		return info, nil
	}
	return "OK", nil
}

// json

func (source *Source) LintJSON(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	body := buffer.Bytes()
	var js json.RawMessage
	e := json.Unmarshal(body, &js)

	if e == nil {
		return "OK", nil
	}

	return fmt.Sprintf("Not valid JSON in `%s`: %s", source.String(), e.Error()), nil
}

// xml

func (source *Source) LintXML(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	decoder := xml.NewDecoder(buffer)
	for {
		err := decoder.Decode(new(interface{}))
		if err != nil && err == io.EOF {
			return "OK", nil
		}
		if err != nil {
			return fmt.Sprintf("Not valid XML in `%s`: %s", source.String(), err.Error()), nil
		}
	}
}

// YAML

func (source *Source) LintYAML(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	decoder := qyaml.NewDecoder(buffer)
	for {
		err := decoder.Decode(new(interface{}))
		if err != nil && err == io.EOF {
			return "OK", nil
		}
		if err != nil {
			return fmt.Sprintf("Not valid YAML in `%s`: %s", source.String(), err.Error()), nil
		}
	}
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
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" {
		info = handleObjects(objs)
		if info == "" {
			info = "OK"
		}
	}

	if info == "OK" && !strings.Contains(preamble, "About") {
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
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" {
		info = handleObjects(objs)
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" && !strings.Contains(preamble, "About") {
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
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" {
		info = handleObjects(objs)
	}
	if info == "OK" && !strings.Contains(preamble, "About") {
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
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" {
		info = handleObjects(objs)
	}
	if info == "OK" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil

}

// Parse MFile
func (source *Source) LintM(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	preamble := "About"
	info = "OK"
	if info == "OK" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return
}

// Parse XFile
func (source *Source) LintX(buffer *bytes.Buffer, warnings bool) (info string, err error) {
	xfile := new(qofile.XFile)
	xfile.Source = source.String()
	xfile.Version = source.Release().String()
	preamble, _, e := xfile.Parse(buffer.Bytes(), true)
	info = "OK"
	if e != nil {
		info = e.Error()
		if info == "" {
			info = "OK"
		}
	}
	if info == "OK" && !strings.Contains(preamble, "About") {
		info = fmt.Sprintf("No `About` in `%s`", source.String())
	}
	return info, nil
}

func handleObjects(objs []qobject.Object) string {
	if len(objs) == 0 {
		return "OK"
	}
	m := make(map[string]bool)
	version := ""
	for _, obj := range objs {
		if version == "" {
			version = obj.Release()
		}
		name := obj.String()
		if m[name] {
			return fmt.Sprintf("Object `%s` occurs more than one in `%s`", name, obj.EditFile())
		}
		err := obj.Lint()
		if err != nil && len(err) > 0 {
			return fmt.Sprintf("Object `%s` has syntax error: `%s`", name, err[0].Error())
		}
	}
	if version == "" {
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
	info := strings.Join(infos, "; ")
	if info == "" {
		info = "OK"
	}
	return info
}
