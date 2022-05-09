package client

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

// LocalFile represents a local file
type LocalFile struct {
	Release  string `json:"release"`
	QPath    string `json:"qpath"`
	DevPath  string `json:"devpath"`
	Project  string `json:"project"`
	Time     string `json:"time"`
	Digest   string `json:"digest"`
	Cu       string `json:"cu"`
	Mu       string `json:"mu"`
	Ct       string `json:"ct"`
	Mt       string `json:"mt"`
	Place    string `json:"-"`
	Sort     string `json:"sort"`
	Priority string `json:"priority"`
	Core     bool   `json:"core"`
}

// Transport of files to B
type Transport struct {
	LocFile LocalFile
	Body    []byte
	Info    []byte
}

type Cargo struct {
	Transports []Transport
	Data       []byte
	Error      []byte
}

func (cargo *Cargo) AddError(e error) {
	if e == nil {
		return
	}
	b := cargo.Error
	var r []interface{}
	if len(b) != 0 {
		var any []interface{}
		er := json.Unmarshal(b, &any)
		if er != nil {
			r = append(r, string(b))
		} else {
			r = append(r, any)
		}
	}
	switch v := e.(type) {
	case qerror.ErrorSlice:
		if len(v) == 0 {
			break
		}
		for _, e := range v {
			if e == nil {
				continue
			}
			var any interface{}
			s := e.Error()
			er := json.Unmarshal([]byte(s), &any)
			if er == nil {
				r = append(r, any)
			} else {
				r = append(r, s)
			}
		}
	default:
		var any interface{}
		s := e.Error()
		er := json.Unmarshal([]byte(s), &any)
		if er == nil {
			r = append(r, any)
		} else {
			r = append(r, s)
		}
	}
	if len(r) == 0 {
		cargo.Error = nil
	} else {
		cargo.Error, _ = json.Marshal(r)
	}
}

// Payload simple instructions
type Payload struct {
	ID         string
	UID        string
	CMD        string
	Origin     string
	Args       []string
	Transports []Transport
	Query      qsource.SQuery
}

func (payload *Payload) GetID() string {
	return payload.ID
}

func (payload *Payload) GetOrigin() string {
	return payload.Origin
}

func (payload *Payload) SetOrigin(origin string) {
	if origin == "" {
		origin = qregistry.Registry["qtechng-type"]
	}
	payload.Origin = origin
}

func (payload *Payload) GetUID() string {
	return payload.UID
}

func (payload *Payload) GetCMD() string {
	return payload.CMD
}

func (payload *Payload) Send(encoder *gob.Encoder) error {
	gob.Register(qerror.ErrorSlice{})
	gob.Register(qerror.QError{})
	return encoder.Encode(*payload)
}

// SendCargo sends data over SSH, back to the initiator
func SendCargo(cargo *Cargo) error {
	gob.Register(qerror.ErrorSlice{})
	gob.Register(qerror.QError{})
	enc := gob.NewEncoder(os.Stdout)
	err := enc.Encode(cargo)

	return err

}

// ReceivePayload receives payload over SSH
func ReceivePayload(wire io.Reader) (ppayload *Payload) {
	gob.Register(qerror.ErrorSlice{})
	gob.Register(qerror.QError{})
	ppayload = &Payload{}
	//all, err := io.ReadAll(wire)
	//log.Fatal("readDataBySSH decoding/0: ", string(all), "\n====\n", err)
	dec := gob.NewDecoder(wire)
	for {
		if err := dec.Decode(ppayload); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("readDataBySSH decoding/1: ", err)
		}
	}
	return ppayload
}

// ReceiveCargo sends data over SSH
func ReceiveCargo(wire *bytes.Buffer) (pcargo *Cargo) {
	gob.Register(qerror.ErrorSlice{})
	gob.Register(qerror.QError{})
	dec := gob.NewDecoder(wire)
	pcargo = new(Cargo)
	for {
		if err := dec.Decode(pcargo); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("readDataBySSH decoding/2: ", err)
		}
	}
	return pcargo
}

// Changed true/file a localfile is changed
func (locfil LocalFile) Changed(place string) bool {
	status := false

	mt, e := qfs.GetMTime(place)
	if e != nil {
		return false
	}
	touch := mt.Format(time.RFC3339)
	if touch != locfil.Time {
		status = true
	}
	return status
}

// Dir represents a directory
type Dir struct {
	Dir   string
	Files map[string]LocalFile
}

// Load loads a directory
func (dir *Dir) Load() {
	if len(dir.Files) != 0 {
		return
	}
	qjson := filepath.Join(dir.Dir, ".qtechng")
	blob, err := os.ReadFile(qjson)
	if err != nil {
		dir.Files = nil
		return
	}
	fis, _, _ := qfs.FilesDirs(dir.Dir)
	isfis := make(map[string]bool)
	for _, fi := range fis {
		isfis[fi.Name()] = true
	}

	files := make(map[string]LocalFile)
	json.Unmarshal(blob, &files)
	change := false
	for base := range files {
		if isfis[base] {
			continue
		}
		delete(files, base)
		change = true
	}
	if change {
		qjson := filepath.Join(dir.Dir, ".qtechng")
		qfs.Store(qjson, files, "qtech")
	}
	dir.Files = files
}

// Get localfile associated with basename
func (dir *Dir) Get(base string) *LocalFile {
	dir.Load()
	locfil, ok := dir.Files[base]
	if !ok {
		return nil
	}
	return &locfil
}

// Add a newcomer to the directory
func (dir *Dir) Add(locfils ...LocalFile) {
	dir.Load()
	ok := false
	if dir.Files == nil {
		dir.Files = make(map[string]LocalFile)
		ok = true
	}
	for _, locfil := range locfils {
		qpath := locfil.QPath
		if qpath == "" || dir.Dir == "" || locfil.Release == "" {
			continue
		}
		_, base := qutil.QPartition(qpath)
		if base == "" || base == ".qtechng" {
			continue
		}
		if !qfs.IsFile(filepath.Join(dir.Dir, base)) {
			continue
		}
		olddigest := dir.Files[base].Digest
		if locfil.Digest == "" {
			locfil.Digest = olddigest
		}
		dir.Files[base] = locfil
		ok = true
	}
	if ok {
		qjson := filepath.Join(dir.Dir, ".qtechng")
		qfs.Store(qjson, dir.Files, "qtech")
		dir.Files = nil
		dir.Load()
	}
}

// Del a file
func (dir *Dir) Del(locfils ...LocalFile) {
	dir.Load()
	if len(dir.Files) == 0 {
		return
	}
	changed := false
	for _, locfil := range locfils {
		qpath := locfil.QPath
		_, base := qutil.QPartition(qpath)
		if base == "" {
			continue
		}
		x, ok := dir.Files[base]
		if !ok {
			continue
		}
		if x.QPath != qpath {
			continue
		}
		delete(dir.Files, base)
		changed = true
	}
	if changed {
		qjson := filepath.Join(dir.Dir, ".qtechng")
		qfs.Store(qjson, dir.Files, "qtech")
		dir.Files = nil
		dir.Load()
	}
}

// List all
func (dir *Dir) List() []LocalFile {
	dir.Load()
	result := make([]LocalFile, len(dir.Files))
	n := 0
	for _, locfil := range dir.Files {
		result[n] = locfil
		n++
	}
	return result
}

// Repository gives all release
func (dir *Dir) Repository() map[string]map[string][]LocalFile {
	dir.Load()
	m := make(map[string]map[string][]LocalFile)
	for _, locfil := range dir.Files {
		r := locfil.Release
		if r == "" {
			continue
		}
		_, ok := m[r]
		if !ok {
			m[r] = make(map[string][]LocalFile)
		}
		qdir, _ := qutil.QPartition(locfil.QPath)
		_, ok = m[r][qdir]
		if !ok {
			m[r][qdir] = make([]LocalFile, 0)
		}

		m[r][qdir] = append(m[r][qdir], locfil)
	}
	if len(m) == 0 {
		return nil
	}
	return m
}

// Find searches/reduces an argument list
func Find(cwd string, files []string, release string, recurse bool, qpatterns []string, onlychanged bool, inlist string, notinlist string, f func(plocfil *LocalFile) bool) (result []*LocalFile, err error) {
	if cwd == "" {
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	if len(files) == 0 {
		files = []string{cwd}
	}

	dirs := make(map[string]map[string]bool)

	for _, fname := range files {
		if fname == "" {
			fname = cwd
		}
		fname = qutil.AbsPath(fname, cwd)
		if !qfs.Exists(fname) {
			continue
		}
		if qfs.IsFile(fname) {
			dir := filepath.Dir(fname)
			base := filepath.Base(fname)
			_, ok := dirs[dir]
			if ok && dirs[dir] == nil {
				continue
			}
			if ok {
				dirs[dir][base] = true
				continue
			}
			if qfs.Exists(filepath.Join(dir, ".qtechng")) {
				dirs[dir] = map[string]bool{base: true}
			}
			continue
		}
		matches, err := qfs.Find(fname, []string{".qtechng"}, recurse, true, false)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			d := filepath.Dir(m)
			dirs[d] = nil
		}
	}

	dirsslice := make([]string, 0)

	for d := range dirs {
		dirsslice = append(dirsslice, d)
	}

	minlist := qutil.FromList(inlist)
	mnotinlist := qutil.FromList(notinlist)

	fn := func(n int) (interface{}, error) {
		dir := dirsslice[n]
		d := new(Dir)
		d.Dir = dir
		d.Load()
		result := make([]*LocalFile, 0)
		for _, locfil := range d.Files {
			if release != "" && locfil.Release != release {
				continue
			}
			qpath := locfil.QPath
			if mnotinlist[qpath] {
				continue
			}
			if len(minlist) != 0 && !minlist[qpath] {
				continue
			}
			_, base := qutil.QPartition(qpath)
			place := filepath.Join(dir, base)

			if dirs[dir] != nil {
				if !dirs[dir][base] {
					continue
				}
			}
			if onlychanged && !locfil.Changed(place) {
				continue
			}
			ok := len(qpatterns) == 0
			for _, qpattern := range qpatterns {
				if !qutil.EMatch(qpattern, qpath) {
					continue
				}
				ok = true
				break
			}
			if !ok {
				continue
			}
			plocfil := new(LocalFile)
			*plocfil = locfil
			plocfil.Place = place

			if f != nil && !f(plocfil) {
				continue
			}
			result = append(result, plocfil)
		}
		return result, nil
	}
	resultlist, _ := qparallel.NMap(len(dirsslice), -1, fn)

	for _, rl := range resultlist {
		result = append(result, rl.([]*LocalFile)...)
	}

	return result, nil
}
