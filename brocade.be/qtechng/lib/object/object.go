package object

import (
	"encoding/json"
	"strings"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Object staat voor een object
type Object interface {
	String() string
	Name() string
	SetName(id string)
	Type() string
	Release() string
	SetRelease(release string)
	EditFile() string
	SetEditFile(editfile string)
	Lineno() string
	SetLineno(lineno string)
	Marshal() ([]byte, error)
	Unmarshal(blob []byte) error
	Loads(blob []byte) error
	Lint() qerror.ErrorSlice
	Format() string
	Deps() []byte
	Replacer(env map[string]string, original string) string
}

// SourceList retrieves the source of a list
func SourceList(objectlist []Object) (sourcemap map[string]string) {
	sourcemap = make(map[string]string)

	fn := func(n int) (interface{}, error) {
		object := objectlist[n]
		oldsource := object.EditFile()
		err := Fetch(object)
		if err != nil {
			return oldsource, nil
		}
		return object.EditFile(), nil
	}

	resultlist, _ := qparallel.NMap(len(objectlist), -1, fn)

	for i, r := range resultlist {
		object := objectlist[i]
		if r == nil {
			continue
		}
		sourcemap[object.String()] = r.(string)
	}
	return
}

// FetchList retrieves all object
func FetchList(objectlist []Object) (objectmap map[string]Object) {
	objectmap = make(map[string]Object)

	fn := func(n int) (interface{}, error) {
		object := objectlist[n]
		err := Fetch(object)
		return object, err
	}

	resultlist, errorlist := qparallel.NMap(len(objectlist), -1, fn)

	for i, result := range resultlist {
		object := result.(Object)
		obj := object.String()
		if errorlist[i] == nil {
			objectmap[obj] = object
			continue
		}
		objectmap[obj] = nil
	}
	return objectmap
}

// Fetch vul een object
func Fetch(object Object) (err error) {
	r := object.Release()
	ty := object.Type()

	if r == "" {
		err := &qerror.QError{
			Ref:    []string{"object.fetch.version.unspecified"},
			File:   object.EditFile(),
			Lineno: 1,
			Object: object.String(),
			Type:   "Error",
			Msg:    []string{"Unspecified version"},
		}
		return err
	}
	rel, e := qserver.Release{}.New(r, true)

	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"object.fetch.version"},
			Version: r,
			File:    object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return qerror.QErrorTune(e, err)
	}

	name := object.String()

	if name == "" {
		err := &qerror.QError{
			Ref:     []string{"object.fetch.name"},
			Version: r,
			File:    object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"Name is empty"},
		}
		return err
	}

	fs := rel.FS("objects", ty)
	h := qutil.Digest([]byte(object.String()))
	dirname := "/" + h[0:2] + "/" + h[2:]
	fname := dirname + "/obj.json"
	content, erro := fs.ReadFile(fname)

	if erro != nil {
		e := &qerror.QError{
			Ref:     []string{"object.fetch.file"},
			Version: r,
			File:    object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"cannot fetch object"},
		}
		return qerror.QErrorTune(erro, e)
	}
	erro = object.Unmarshal(content)
	if erro != nil {
		e := &qerror.QError{
			Ref:     []string{"object.fetch.unmarshal"},
			Version: r,
			File:    object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"cannot unmarshal object"},
		}
		return qerror.QErrorTune(erro, e)
	}
	return nil
}

// StoreList stores a list of object's
func StoreList(objectlist []Object) (changedmap map[string]bool, errorlist []error) {

	changedmap = make(map[string]bool)

	fn := func(n int) (interface{}, error) {
		object := objectlist[n]
		changed, err := Store(object)
		return changed, err
	}

	resultlist, errorlist := qparallel.NMap(len(objectlist), -1, fn)

	for i, r := range resultlist {
		changed := false
		object := objectlist[i]
		if r != nil {
			changed = r.(bool)
		}
		if changed {
			changedmap[object.String()] = true
		}
		err := errorlist[i]
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"objectlist.store.file"},
				Version: object.Release(),
				File:    object.EditFile(),
				Lineno:  1,
				Object:  object.String(),
				Type:    "Error",
				Msg:     []string{"Cannot store object"},
			}
			errorlist = append(errorlist, e)
		} else {
			errorlist = append(errorlist, nil)
		}
	}
	return
}

// Store Opslag van een object
func Store(object Object) (changed bool, err error) {
	name := object.String()
	r := object.Release()
	if name == "" {
		err := &qerror.QError{
			Ref:     []string{"object.store.name"},
			Version: r,
			File:    object.EditFile(),
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"Name is empty"},
		}
		return false, err
	}

	editfile := object.EditFile()
	if editfile == "" {
		err := &qerror.QError{
			Ref:     []string{"object.store.editfile"},
			Version: r,
			File:    editfile,
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"Missing editfile"},
		}
		return false, err
	}
	if r == "" {
		err := &qerror.QError{
			Ref:     []string{"object.store.version1"},
			Version: r,
			File:    editfile,
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"Unspecified version"},
		}
		return false, err
	}
	rel, e := qserver.Release{}.New(r, false)
	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"object.store.version2"},
			Version: r,
			File:    editfile,
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return false, qerror.QErrorTune(e, err)
	}
	fs := rel.FS("objects", object.Type())
	h := qutil.Digest([]byte(name))
	dirname := "/" + h[0:2] + "/" + h[2:]
	fs.MkdirAll(dirname, 0770)
	changed, before, actual, err := fs.Store(dirname+"/obj.json", object, "")

	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"object.store.file"},
			Version: r,
			File:    editfile,
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"cannot store object"},
		}
		return false, qerror.QErrorTune(err, e)
	}
	if !changed {
		return false, nil
	}
	Link(r, editfile, object)
	oactual := object.Deps()

	if len(before) != 0 {
		object.Unmarshal(before)
		before = object.Deps()
		object.Unmarshal(actual)
	}

	StoreLinks(r, name, before, oactual)

	return true, nil
}

// Link links objects with a source
func Link(r string, name string, object interface{}) error {
	obj := ""
	switch v := object.(type) {
	case string:
		obj = v
	case []byte:
		obj = string(v)
	case Object:
		obj = v.String()
		if name == "" {
			name = v.EditFile()
		}
		if r == "" {
			r = v.Release()
		}
	}

	version, e := qserver.Release{}.New(r, false)
	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"object.link.version"},
			Version: r,
			File:    name,
			Lineno:  1,
			Object:  obj,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return qerror.QErrorTune(e, err)
	}
	data := map[string]string{"source": name}
	fs := version.FS("/objects")
	digest := qutil.Digest([]byte(name))
	dir2 := "/" + digest[:2] + "/" + digest[2:] + ".dep"
	if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
		parts := strings.SplitN(obj, "_", 3)
		obj = "l4_" + parts[2]
	}
	parts := strings.SplitN(obj, "_", 2)
	mode := parts[0]
	digest = qutil.Digest([]byte(obj))
	fname := "/" + mode + "/" + digest[:2] + "/" + digest[2:] + dir2
	data["object"] = obj
	fs.Store(fname, data, "")
	return nil
}

// StoreLinks stores the links from object to where the object is used
func StoreLinks(r string, name string, before []byte, actual []byte) {
	obefore := make(map[string]bool)
	oactual := make(map[string]bool)

	if len(actual) != 0 {
		odd := true
		for _, bobj := range qutil.ObjectSplitter(actual) {
			odd = !odd
			if odd {
				obj := string(bobj)
				if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
					parts := strings.SplitN(obj, "_", 3)
					obj = "l4_" + parts[2]
				}
				oactual[obj] = true
			}
		}
	}

	if len(before) != 0 {
		odd := true
		for _, bobj := range qutil.ObjectSplitter(before) {
			odd = !odd
			if odd {
				obj := string(bobj)
				if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
					parts := strings.SplitN(obj, "_", 3)
					obj = "l4_" + parts[2]
				}
				obefore[obj] = true
			}
		}
	}

	for obj := range oactual {

		if !obefore[obj] {
			Link(r, name, obj)
		}
	}

	for obj := range obefore {
		if !oactual[obj] {
			UnLink(r, name, obj)
		}
	}
}

// UnLink links objects with a source
func UnLink(r string, name string, object interface{}) error {
	obj := ""
	switch v := object.(type) {
	case string:
		obj = v
	case []byte:
		obj = string(v)
	case Object:
		obj = v.String()
		if name == "" {
			name = v.EditFile()
		}
		if r == "" {
			r = v.Release()
		}
	}
	if len(obj) < 4 {
		err := &qerror.QError{
			Ref:     []string{"object.unlink.object"},
			Version: r,
			File:    name,
			Lineno:  1,
			Object:  obj,
			Type:    "Error",
			Msg:     []string{"Not a valid object"},
		}
		return err
	}
	version, e := qserver.Release{}.New(r, false)
	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"object.unlink.version"},
			Version: r,
			File:    name,
			Lineno:  1,
			Object:  obj,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return qerror.QErrorTune(e, err)
	}
	mode := obj[:2]
	fs := version.FS("objects", mode)
	digest := qutil.Digest([]byte(name))
	dir2 := "/" + digest[:2] + "/" + digest[2:] + ".dep"
	if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
		parts := strings.SplitN(obj, "_", 3)
		obj = "l4_" + parts[2]
	}
	digest = qutil.Digest([]byte(obj))
	fname := "/" + digest[:2] + "/" + digest[2:] + dir2
	fs.Waste(fname)
	return nil
}

// GetDependencies geeft de bestanden die afhankelijk zijn van een object
func GetDependencies(version *qserver.Release, objs ...string) map[string][]string {

	fs := version.FS("/objects")
	fn := func(n int) (interface{}, error) {
		obj := objs[n]
		if strings.HasPrefix(obj, "l4") && strings.Count(obj, "_") == 2 {
			parts := strings.SplitN(obj, "_", 3)
			obj = "l4_" + parts[2]
		}
		parts := strings.SplitN(obj, "_", 2)
		mode := parts[0]
		digest := qutil.Digest([]byte(obj))
		startdir := "/" + mode + "/" + digest[:2] + "/" + digest[2:]
		matches := fs.Glob(startdir, []string{"*.dep"}, true)
		dep := make(map[string]string)
		dono := make([]string, 0)
		for _, match := range matches {
			content, err := fs.ReadFile(match)
			if err != nil {
				continue
			}
			err = json.Unmarshal(content, &dep)
			if err != nil {
				continue
			}
			dono = append(dono, dep["source"])
		}
		return dono, nil
	}

	rlist, _ := qparallel.NMap(len(objs), -1, fn)
	result := make(map[string][]string)

	for n, obj := range objs {
		dono := rlist[n].([]string)
		if len(dono) == 0 {
			result[obj] = nil
		} else {
			result[obj] = dono
		}
	}
	return result
}
