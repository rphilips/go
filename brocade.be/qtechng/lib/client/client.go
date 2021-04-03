package client

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	qfnmatch "brocade.be/base/fnmatch"
	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

// LocalFile represents a local file
type LocalFile struct {
	Release string `json:"release"`
	QPath   string `json:"qpath"`
	Project string `json:"project"`
	Time    string `json:"time"`
	Digest  string `json:"digest"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
	Place   string `json:"-"`
}

// Transport of files to B
type Transport struct {
	LocFile LocalFile
	Body    []byte
}

type Cargo struct {
	Transports []Transport
	Buffer     bytes.Buffer
	Error      error
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
	for {
		if err := dec.Decode(pcargo); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("readdDataBySSH decoding/2: ", err)
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
	qjson := path.Join(dir.Dir, ".qtechng")
	blob, err := os.ReadFile(qjson)
	if err != nil {
		dir.Files = nil
		return
	}
	fis, _, err := qfs.FilesDirs(dir.Dir)
	isfis := make(map[string]bool)
	for _, fi := range fis {
		isfis[fi.Name()] = true
	}

	files := make(map[string]LocalFile)
	json.Unmarshal(blob, &files)
	change := false
	for base, _ := range files {
		if isfis[base] {
			continue
		}
		delete(files, base)
		change = true
	}
	if change {
		qjson := path.Join(dir.Dir, ".qtechng")
		qfs.Store(qjson, files, "")
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
	if dir.Files == nil {
		dir.Files = make(map[string]LocalFile)
	}
	ok := false
	for _, locfil := range locfils {
		qpath := locfil.QPath
		if qpath == "" || dir.Dir == "" || locfil.Release == "" {
			continue
		}
		_, base := qutil.QPartition(qpath)
		if base == "" || base == ".qtechng" {
			continue
		}
		if !qfs.IsFile(path.Join(dir.Dir, base)) {
			continue
		}
		dir.Files[base] = locfil
		ok = true
	}
	if ok {
		qjson := path.Join(dir.Dir, ".qtechng")
		qfs.Store(qjson, dir.Files, "")
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
		x := dir.Files[base]
		if x.QPath == qpath {
			delete(dir.Files, base)
			changed = true
		}
	}
	if changed {
		qjson := path.Join(dir.Dir, ".qtechng")
		dir.Files = nil
		qfs.Store(qjson, dir, "")
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
func Find(cwd string, files []string, release string, recurse bool, qpattern []string, onlychanged bool) (result []*LocalFile, err error) {
	find := false
	if len(files) == 0 {
		find = true
		files, err = qfs.Find(cwd, nil, recurse, true, false)
		if err != nil {
			err := &qerror.QError{
				Ref:  []string{"client.find.io"},
				Type: "Error",
				Msg:  []string{"Cannot find files: " + err.Error()},
			}
			return nil, err
		}
	}
	done := make(map[string]bool)
	qpaths := make(map[string]string)

	result = make([]*LocalFile, 0)
	errlist := make([]error, 0)

	sort.Strings(files)
	d := new(Dir)
	pdir := "/"
	for _, file := range files {
		if file == "" {
			continue
		}
		if done[file] {
			continue
		}
		done[file] = true
		place := file
		if !path.IsAbs(file) {
			place = path.Join(cwd, place)
		}
		dir := filepath.Dir(place)
		base := filepath.Base(place)
		if pdir != dir {
			pdir = dir
			d = new(Dir)
			d.Dir = dir
		}
		plocfil := d.Get(base)
		if plocfil == nil {
			if onlychanged {
				continue
			}
			if !find {
				err := &qerror.QError{
					Ref:  []string{"client.find.get"},
					Type: "Error",
					File: place,
					Msg:  []string{"`" + base + "` does not exists in QtechNG"},
				}
				errlist = append(errlist, err)
			}
			continue
		}
		if onlychanged && !plocfil.Changed(place) {
			continue
		}
		if release != "" {
			rok := qfnmatch.Match(release, plocfil.Release)

			if !rok && !find {
				err := &qerror.QError{
					Ref:  []string{"client.find.version"},
					Type: "Error",
					Msg:  []string{"`" + file + "` does not match version"},
				}
				errlist = append(errlist, err)
			}
			if !rok {
				continue
			}
		}
		if len(qpattern) != 0 {
			qok := false
			for _, qpat := range qpattern {
				qok = qfnmatch.Match(qpat, plocfil.QPath)
				if qok {
					break
				}
			}
			if !qok && !find {
				err := &qerror.QError{
					Ref:  []string{"client.find.qpath"},
					Type: "Error",
					Msg:  []string{"`" + file + "` does not match pattern"},
				}
				errlist = append(errlist, err)
			}
			if !qok {
				continue
			}
		}
		plocfil.Place = place

		qpath := plocfil.QPath

		ofile, qok := qpaths[qpath]

		if qok {
			err := &qerror.QError{
				Ref:  []string{"client.find.doubleqpath"},
				Type: "Error",
				Msg:  []string{"`" + file + "` and `" + ofile + "` refer both to `" + qpath},
			}
			errlist = append(errlist, err)
			continue
		}
		qpaths[qpath] = file
		result = append(result, plocfil)
	}

	if len(result) == 0 {
		result = nil
	}
	if len(errlist) == 0 {
		return result, nil
	}
	return result, qerror.ErrorSlice(errlist)
}
