package source

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Rename renames - with version r - all files in the map
func Rename(r string, qpaths map[string]string, overwrite bool) (err error) {
	// versions

	rversion, err := qserver.Release{}.New(r, true)

	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"source.rename.version"},
			Version: r,
			Msg:     []string{"Cannot instantiate version"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}

	err = TestForRenames(rversion, qpaths, overwrite)
	if err != nil {
		return err
	}

	wversion, err := qserver.Release{}.New(r, false)

	// find objects
	//objfiles := objWithSource(rversion, qpaths)

	// write sources
	fs, _ := wversion.SourcePlace("/")
	//written := make(map[string]bool)
	for qpath, target := range qpaths {
		data, e := fs.ReadFile(qpath)
		if e != nil {
			continue
		}
		fs.Store(target, data, "")
		fs.RemoveAll(qpath)

		fm, pm := wversion.MetaPlace(qpath)
		meta, e := fm.ReadFile(pm)
		if e != nil {
			m := make(map[string]string)
			e := json.Unmarshal(meta, &m)
			if e != nil {
				m["source"] = target
			}
			meta, _ = json.Marshal(m)
			fm.RemoveAll(pm)
		}
		fm, pm = wversion.MetaPlace(target)
		fm.Store(pm, meta, "")
		wversion.UniqueStore(target)
		wversion.UniqueUnlink(qpath)
		chgWithSource(rversion, qpaths)
	}

	// rename source
	// remove superflous
	// rename meta
	// remove superfluous
	// replaces objects
	// replaces dependencies
	// reports

	return
}

func TestForRenames(rversion *qserver.Release, qpaths map[string]string, overwrite bool) (err error) {
	targets := make(map[string]string)
	for qpath, target := range qpaths {
		targets[target] = qpath
	}
	for qpath := range qpaths {
		opath := targets[qpath]
		if opath != "" {
			return fmt.Errorf("`%s`, renamed to `%s`, will first be overwritten by `%s`", qpath, qpaths[qpath], opath)
		}
	}

	removed := make([]string, 0)
	added := make([]string, 0)
	for qpath, target := range qpaths {
		removed = append(removed, qpath)
		added = append(added, target)
	}
	r := rversion.String()

	fs, _ := rversion.SourcePlace("/")
	for qpath, target := range qpaths {
		ok, _ := fs.DirExists(target)
		if ok {
			return fmt.Errorf("`%s`, renamed to `%s`, is a directory", qpath, target)
		}
		if overwrite {
			continue
		}
		ok, _ = fs.Exists(target)
		if ok {
			return fmt.Errorf("`%s`, renamed to `%s`, overwrites existing file", qpath, target)
		}
	}

	newconfigs := make([]string, 0)
	except := make(map[string]bool)

	for qpath, target := range qpaths {
		if strings.HasSuffix(target, "/brocade.json") && strings.HasSuffix(qpath, "/brocade.json") {
			psource, e := Source{}.New(r, qpath, true)
			if e != nil {
				continue
			}
			natures := psource.Natures()
			if !natures["config"] {
				continue
			}
			newconfigs = append(newconfigs, target)
			continue
		}
		ext1 := path.Ext(qpath)
		ext2 := path.Ext(target)
		if ext1 != ext2 {
			continue
		}
		if ext1 == ".d" || ext1 == ".l" || ext1 == ".i" {
			psource, e := Source{}.New(r, qpath, true)
			if e != nil {
				continue
			}
			natures := psource.Natures()
			if !natures["objectfile"] {
				continue
			}
			except[qpath] = true
			continue
		}
	}
	err = TestForWasteList(r, removed, added, newconfigs, except)
	if err != nil {
		return err
	}

	// test on adding files

	mapprojs := make(map[string]bool)

	for _, cfg := range newconfigs {
		p, _ := qutil.QPartition(cfg)
		mapprojs[p] = true
	}
	for _, target := range targets {
		p, _ := qutil.QPartition(target)
		if mapprojs[p] {
			continue
		}
		_, e := Source{}.New(r, target, true)
		if e != nil {
			return e
		}
		mapprojs[p] = true
	}
	return nil
}

func chgWithSource(rversion *qserver.Release, qpaths map[string]string) []string {
	tys := make([]string, 0)
	for qpath := range qpaths {
		ext := path.Ext(qpath)
		ty := ""
		switch ext {
		case ".l":
			ty = "l4"
		case ".d":
			ty = "m4"
		case ".i":
			ty = "i4"
		}
		if ty == "" {
			continue
		}
		for _, x := range tys {
			if x == ty {
				ty = ""
				break
			}
		}
		if ty == "" {
			continue
		}
		tys = append(tys, ty)
	}
	if len(tys) == 0 {
		return nil
	}

	fn1 := func(n int) (interface{}, error) {
		ty := tys[n]
		fs, _ := rversion.ObjectPlace(ty + "_xyz")
		dir, _ := fs.RealPath("/")
		all, _ := qfs.Find(dir, []string{"obj.json"}, true, true, false)
		if len(all) == 0 {
			return nil, nil
		}
		fn2 := func(n int) (interface{}, error) {
			fname := all[n]
			m := make(map[string]interface{})
			body, e := qfs.Fetch(fname)
			if e != nil {
				return "", nil
			}
			if len(body) == 0 {
				return "", nil
			}
			e = json.Unmarshal(body, &m)
			if e != nil {
				return "", nil
			}
			source, ok := m["source"]
			if !ok {
				return "", nil
			}
			s := source.(string)
			if qpaths[s] == "" {
				return "", nil
			}
			m["source"] = qpaths[s]
			body, _ = json.Marshal(m)
			qfs.Store(fname, body, "qtech")
			return fname, nil
		}
		resultlist, _ := qparallel.NMap(len(all), -1, fn2)
		result := make([]string, 0)
		for _, r := range resultlist {
			s := r.(string)
			if s == "" {
				continue
			}
			result = append(result, s)
		}
		return result, nil
	}

	files := make([]string, 0)
	resultlist, _ := qparallel.NMap(len(tys), -1, fn1)
	for _, rl := range resultlist {
		if rl == nil {
			continue
		}
		for _, ol := range rl.([]string) {
			if ol == "" {
				continue
			}
			files = append(files, ol)
		}
	}
	return files

}
