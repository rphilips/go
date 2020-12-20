package objfile

import (
	"fmt"
	"path"
	"strings"

	qfs "brocade.be/base/fs"
	qobject "brocade.be/qtech/object"
	qutil "brocade.be/qtech/util"
	qerror "brocade.be/qtechng/error"
)

// ExtractTexts returns the texts for the objects
func ExtractTexts(file string) (about string, texts []qutil.EditFile, err error) {
	absfile, erro := qfs.AbsPath(file)
	if erro != nil {
		absfile = file
	}
	data, err := qfs.Fetch(file)
	if err != nil {
		e := &qerror.QError{
			Ref:  []string{"objectfile.fetch"},
			File: absfile,
			Msg:  []string{fmt.Sprintf("Fail to read `%s`", absfile)},
		}
		err = qerror.QErrorTune(err, e)
		return
	}

	leader := ""
	ext := path.Ext(file)

	switch ext {
	case ".d":
		leader = "macro"
	case ".l":
		leader = "lgcode"
	case ".i":
		leader = "include"
	default:
		leader = ""
	}

	if leader == "" {
		err = &qerror.QError{
			Ref:  []string{"objectfile.ext"},
			File: absfile,
			Msg:  []string{fmt.Sprintf("File `%s` has the wrong extension", absfile)},
		}
		return
	}

	efile := qutil.NewEditFile(string(data), 0)
	_, ebody := efile.About()

	var text qutil.EditFile

	for _, line := range ebody {
		l := line.Text
		if strings.HasPrefix(l, leader) {
			if len(text) != 0 {
				texts = append(texts, text)
			}
			text = qutil.EditFile{line}
			continue
		}
		if len(text) != 0 {
			text = append(text, line)
		}
	}
	if len(text) != 0 {
		texts = append(texts, text)
	}
	return
}

// Parse file and return the objects
func Parse(file string) (objs []qobject.Object, err error) {
	absfile, erro := qfs.AbsPath(file)
	if erro != nil {
		absfile = file
	}

	errmap := qerror.NewErrorMap()
	infmap := make(map[string]string)

	about, texts, err := ExtractTexts(file)
	if err != nil {
		errmap.AddError("Extract", err)
		return
	}
	if erro != nil {
		absfile = file
	}
	if about == "" {
		err = &qerror.QError{
			Ref:  []string{"objectfile.parse.about"},
			File: absfile,
			Msg:  []string{fmt.Sprintf("File `%s` has no `about`", absfile)},
		}
		errmap.AddError("About", err)
		return
	}

	ext := path.Ext(file)

	for _, text := range texts {
		var obj qobject.Object
		switch ext {
		case ".d":
			obj = new(qobject.Macro)
		case ".l":
			obj = new(qobject.Lgcode)
		case ".i":
			obj = new(qobject.Include)
		}
		erro := obj.Parse(text, infmap)
		if erro != nil {
			e := &qerror.QError{
				Ref:  []string{"objectfile.parse.obj"},
				File: absfile,
				Msg:  []string{"Fail to parse"},
			}
			err = qerror.QErrorTune(erro, e)
			errmap.AddError(text[0].Text, err)
			continue
		}
		objs = append(objs, obj)
	}
	if len(errmap) == 0 {
		err = nil
		return
	}
	err = errmap
	return
}
