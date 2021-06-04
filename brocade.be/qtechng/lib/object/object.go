package object

import (
	"encoding/json"
	"fmt"
	"sort"
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
	MarshalJSON() ([]byte, error)
	Unmarshal(blob []byte) error
	Loads(blob []byte) error
	Lint() qerror.ErrorSlice
	Format() string
	Deps() []byte
	Replacer(env map[string]string, original string) string
}

// Uber object
type Uber struct {
	ID      string `json:"id"`   // Identificatie
	Body    []byte `json:"meta"` // Meta
	UsedBy  []string
	Source  string `json:"source"` // Editfile
	Line    string `json:"-"`      // Lijnnummer
	Version string `json:"-"`      // Version
}

// String
func (uber *Uber) String() string {
	return uber.ID
}

// Name of uber
func (uber *Uber) Name() string {
	return uber.ID
}

// SetName of uber
func (uber *Uber) SetName(id string) {
	uber.ID = id
}

// Type of uber
func (uber *Uber) Type() string {
	x := uber.String()
	k := strings.Index(x, "_")
	if k < 1 {
		return ""
	}
	return x[:k]
}

// Release of uber
func (uber *Uber) Release() string {
	return uber.Version
}

// SetRelease of uber
func (uber *Uber) SetRelease(version string) {
	uber.Version = version
}

// EditFile of uber
func (uber *Uber) EditFile() string {
	return uber.Source
}

// SetEditFile of uber
func (uber *Uber) SetEditFile(source string) {
	uber.Source = source
}

// Lineno of macro
func (uber *Uber) Lineno() string {
	return uber.Line
}

// SetLineno of macro
func (uber *Uber) SetLineno(lineno string) {
	uber.Line = lineno
}

// Marshal of uber
func (uber *Uber) Marshal() ([]byte, error) {
	return uber.MarshalJSON()
}

// MarshalJSON of uber
func (uber *Uber) MarshalJSON() ([]byte, error) {
	v := new(interface{})
	json.Unmarshal(uber.Body, v)
	vv := make(map[string]interface{})
	vv["definition"] = v
	usedby := uber.UsedBy
	if len(usedby) != 0 {
		sort.Strings(usedby)
	}
	vv["usedby"] = usedby
	return json.MarshalIndent(vv, "", "    ")
}

// Unmarshal of uber
func (uber *Uber) Unmarshal(blob []byte) error {
	return nil
}

// Loads from blob
func (uber *Uber) Loads(blob []byte) error {
	return nil
}

// Deps fake
func (uber *Uber) Deps() []byte {
	return nil
}

// Format fake
func (uber *Uber) Format() string {
	return ""
}

// Lint fake
func (uber *Uber) Lint() (errslice qerror.ErrorSlice) {
	return nil
}

// Replacer fake
func (uber *Uber) Replacer(env map[string]string, original string) string {
	return ""
}

// MakeObjectList makes a list of Objects starting with string
func MakeObjectList(r string, objstr []string) (objectlist []Object) {

	for _, obj := range objstr {
		prefix := ""
		code := ""
		k := strings.Index(obj, "_")
		if k > 1 {
			prefix = obj[:k]
			code = obj[k+1:]
		}
		if prefix == "l4" {
			if strings.Count(code, "_") == 1 {
				parts := strings.SplitN(code, "_", 2)
				obj = "l4_" + parts[1]
			}
		}
		uber := new(Uber)
		uber.SetName(obj)
		uber.SetRelease(r)
		objectlist = append(objectlist, uber)
	}
	return
}

// InfoObjectList maak een structuur aan met object informatie
func InfoObjectList(r string, objstr []string) (objectmap map[string]Object) {
	uberlist := MakeObjectList(r, objstr)
	return FetchList(uberlist)
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

	if r == "" {
		err := &qerror.QError{
			Ref:    []string{"object.fetch.version.unspecified"},
			QPath:  object.EditFile(),
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
			QPath:   object.EditFile(),
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
			QPath:   object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"Name is empty"},
		}
		return err
	}

	fs, fname := rel.ObjectPlace(object.String())
	content, erro := fs.ReadFile(fname)
	if erro != nil {
		e := &qerror.QError{
			Ref:     []string{"object.fetch.file"},
			Version: r,
			QPath:   object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"cannot fetch object"},
		}
		return qerror.QErrorTune(erro, e)
	}
	switch v := object.(type) {
	case *Uber:
		v.Body = content
		deps, _ := GetDependenciesDeep(rel, v.ID)
		v.UsedBy = deps[v.ID]
		u := make(map[string]string)
		json.Unmarshal(content, &u)
		v.Source = u["source"]

	default:
		erro = object.Unmarshal(content)
	}
	erro = object.Unmarshal(content)
	if erro != nil {
		e := &qerror.QError{
			Ref:     []string{"object.fetch.unmarshal"},
			Version: r,
			QPath:   object.EditFile(),
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
				QPath:   object.EditFile(),
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
			QPath:   object.EditFile(),
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
			QPath:   editfile,
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
			QPath:   editfile,
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
			QPath:   editfile,
			Lineno:  1,
			Object:  name,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return false, qerror.QErrorTune(e, err)
	}
	changed, before, actual, err := rel.ObjectStore(name, object)

	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"object.store.file"},
			Version: r,
			QPath:   editfile,
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
			QPath:   name,
			Lineno:  1,
			Object:  obj,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return qerror.QErrorTune(e, err)
	}
	data := map[string]string{"source": name}
	fs, fname := version.ObjectDepPlace(obj, name)
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
			QPath:   name,
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
			QPath:   name,
			Lineno:  1,
			Object:  obj,
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return qerror.QErrorTune(e, err)
	}
	fs, fname := version.ObjectDepPlace(obj, name)
	fs.Waste(fname)
	return nil
}

// GetDependenciesDeep geeft de bestanden die afhankelijk zijn van een aantal objecten
func GetDependenciesDeep(version *qserver.Release, objs ...string) (result map[string][]string, err error) {
	result = make(map[string][]string)
	iresult := make(map[string]map[string]bool)
	todo := append([]string{}, objs...)
	for {
		newobjs := make([]string, 0)
		for _, obj := range todo {
			if iresult[obj] != nil {
				continue
			}
			newobjs = append(newobjs, obj)
		}
		if len(newobjs) == 0 {
			break
		}
		todo = []string{}
		between := GetDependencies(version, newobjs...)
		for _, obj := range newobjs {
			if iresult[obj] == nil {
				iresult[obj] = make(map[string]bool)
			}
			deps := between[obj]
			for _, d := range deps {
				// if d == obj {
				// 	err := &qerror.QError{
				// 		Ref:    []string{"cyclic.equal"},
				// 		Object: obj,
				// 		Type:   "Error",
				// 		Msg:    []string{fmt.Sprintf("`%s` has cyclic dependency", obj)},
				// 	}
				// 	return nil, err
				// }
				iresult[obj][d] = true
				if !strings.HasPrefix(d, "/") {
					todo = append(todo, d)
				}
			}
		}
	}

	run := true
	for run {
		run = false
		for obj, mdeps := range iresult {
			if len(mdeps) == 0 {
				continue
			}
			if mdeps[obj] {
				err := &qerror.QError{
					Ref:    []string{"cyclic.equal"},
					Object: obj,
					Type:   "Error",
					Msg:    []string{fmt.Sprintf("`%s` has cyclic dependency", obj)},
				}
				return nil, err
			}
			add := make([]string, 0)
			for d := range mdeps {
				if !strings.HasPrefix(d, "/") {
					mc := iresult[d]
					if len(mc) == 0 {
						continue
					}
					for k := range mc {
						add = append(add, k)
					}
				}
			}
			if len(add) != 0 {
				for _, d := range add {
					if !mdeps[d] {
						mdeps[d] = true
						run = true
					}
				}
			}
			iresult[obj] = mdeps
		}
	}

	for _, obj := range objs {
		result[obj] = make([]string, len(iresult[obj]))
		i := -1
		for k := range iresult[obj] {
			i++
			result[obj][i] = k
		}
	}
	return result, nil
}

// GetDependencies retrieves a map pointin for each object the things dependent on that object
func GetDependencies(version *qserver.Release, objs ...string) (result map[string][]string) {
	fn := func(n int) (interface{}, error) {
		obj := objs[n]
		fs, fname := version.ObjectPlace(obj)
		startdir := strings.TrimSuffix(fname, "/obj.json")
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
	result = make(map[string][]string)
	rlist, _ := qparallel.NMap(len(objs), -1, fn)
	for n, obj := range objs {
		dono := rlist[n].([]string)
		result[obj] = dono
	}
	return
}
