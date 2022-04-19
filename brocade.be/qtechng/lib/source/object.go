package source

import (
	"errors"
	"strings"

	bfs "io/fs"

	qfnmatch "brocade.be/base/fnmatch"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qserver "brocade.be/qtechng/lib/server"
)

// Waste vernietig een object
func Waste(object qobject.Object) (changed bool, err error) {
	r := object.Release()
	if r == "" {
		err := &qerror.QError{
			Ref:    []string{"object.waste.version1"},
			QPath:  object.EditFile(),
			Lineno: 1,
			Object: object.String(),
			Type:   "Error",
			Msg:    []string{"Unspecified version"},
		}
		return false, err
	}
	rel, e := qserver.Release{}.New(r, false)

	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"object.waste.version2"},
			Version: r,
			QPath:   object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"Cannot find version on disk"},
		}
		return false, qerror.QErrorTune(e, err)
	}

	rest := Deleteable(object)[object.String()]
	if rest != nil {
		e := &qerror.QError{
			Ref:     []string{"object.waste.deps"},
			Version: r,
			QPath:   object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"Others depend on this object: `" + strings.Join(rest, ", ") + "`"},
		}
		return false, e
	}

	fs, fname := rel.ObjectPlace(object.String())
	changed, err = fs.Waste(fname)

	if err != nil && !errors.Is(err, bfs.ErrNotExist) {
		e := &qerror.QError{
			Ref:     []string{"object.waste.file"},
			Version: r,
			QPath:   object.EditFile(),
			Lineno:  1,
			Object:  object.String(),
			Type:    "Error",
			Msg:     []string{"cannot delete object"},
		}
		return false, qerror.QErrorTune(err, e)
	}
	return changed, nil
}

// WasteObjList deletes a list of object's
func WasteObjList(objectlist []qobject.Object) (changedlist []bool, errorlist []error) {

	fn := func(n int) (interface{}, error) {
		object := objectlist[n]
		changed, err := Waste(object)
		return changed, err
	}

	resultlist, errorlist := qparallel.NMap(len(objectlist), -1, fn)

	for i, r := range resultlist {
		changed := false
		object := objectlist[i]
		if r != nil {
			changed = r.(bool)
		}
		changedlist = append(changedlist, changed)
		err := errorlist[i]
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"objectlist.waste.file"},
				Version: object.Release(),
				QPath:   object.EditFile(),
				Lineno:  1,
				Object:  object.String(),
				Type:    "Error",
				Msg:     []string{"cannot waste object"},
			}
			errorlist = append(errorlist, e)
		} else {
			errorlist = append(errorlist, nil)
		}
	}
	return
}

// Deleteable returns if an object can be deleted
func Deleteable(objects ...qobject.Object) map[string][]string {
	if len(objects) == 0 {
		return nil
	}
	r := objects[0].Release()
	version, e := qserver.Release{}.New(r, true)
	if e != nil {
		return nil
	}
	edfiles := make(map[string]string)
	objs := make([]string, len(objects))
	for i, object := range objects {
		objs[i] = object.String()
		edfiles[object.String()] = object.EditFile()
	}

	deps := qobject.GetDependencies(version, objs...)

	if len(deps) == 0 {
		return nil
	}

	result := make(map[string][]string)
	for obj, value := range deps {
		if value == nil {
			result[obj] = nil
			continue
		}
		found := make([]string, 0)
		editfile := edfiles[obj]
		for _, dep := range deps[obj] {
			if !strings.HasPrefix(dep, "/") {
				found = append(found, dep)
			}
			if dep == editfile {
				continue
			}
			source, err := Source{}.New(r, dep, true)
			if err != nil {
				continue
			}
			natures := source.Natures()
			if !natures["text"] {
				continue
			}
			project := source.Project()
			config, err := project.LoadConfig()
			if (err != nil) || (len(config.ObjectsNotReplaced[obj]) == 0) {
				found = append(found, dep)
				continue
			}
			rel := source.Rel()
			ok := false
			for _, pth := range config.ObjectsNotReplaced[obj] {
				ok = qfnmatch.Match(pth, rel)
				if ok {
					break
				}
			}
			if ok {
				continue
			}
			found = append(found, dep)
		}
		if len(found) == 0 {
			result[obj] = nil
			continue
		} else {
			result[obj] = found
		}
	}
	return result
}
