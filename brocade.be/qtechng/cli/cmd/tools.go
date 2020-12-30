package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qinstall "brocade.be/qtechng/lib/install"
	qmeta "brocade.be/qtechng/lib/meta"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

type lister struct {
	Release string `json:"version"`
	QPath   string `json:"qpath"`
	Project string `json:"project"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
}

type storer struct {
	Release string `json:"version"`
	Qpath   string `json:"qpath"`
	Project string `json:"project"`
	Changed bool   `json:"changed"`
	Time    string `json:"time"`
	Digest  string `json:"digest"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
	Place   string `json:"file"`
}

func fetchData(args []string, filesinproject bool, qdirs []string, mumps bool) (pcargo *qclient.Cargo, err error) {
	if len(args) != 0 {
		if len(Fpattern) == 0 {
			Fpattern = args
		} else {
			for _, arg := range args {
				ok := len(Fpattern) == 0
				for _, p := range Fpattern {
					if p == arg {
						ok = true
						break
					}
					if qutil.EMatch(p, arg) {
						ok = true
						break
					}
				}
				if ok {
					Fpattern = append(Fpattern, arg)
				}
			}
		}
	}

	squery := qsource.SQuery{
		Release:        Fversion,
		CmpRelease:     Fcmpversion,
		QDirs:          qdirs,
		Patterns:       Fpattern,
		FilesInProject: filesinproject,
		Natures:        Fnature,
		Cu:             Fcu,
		Mu:             Fmu,
		CtBefore:       Fctbefore,
		CtAfter:        Fctafter,
		MtBefore:       Fmtbefore,
		MtAfter:        Fmtafter,
		ToLower:        Ftolower,
		Regexp:         Fregexp,
		PerLine:        Fperline,
		SmartCase:      !Fsmartcaseoff,
		Contains:       Fneedle,
		Mumps:          mumps,
	}
	Fpayload = &qclient.Payload{
		ID:     "Once",
		UID:    FUID,
		CMD:    "qtechng",
		Origin: QtechType,
		Args:   os.Args[1:],
		Query:  squery,
	}
	pcargo = &qclient.Cargo{}
	if !strings.ContainsRune(QtechType, 'B') && !strings.ContainsRune(QtechType, 'P') {
		whowhere := FUID + "@" + qregistry.Registry["qtechng-server"]
		catchOut, catchErr, err := qssh.SSHcmd(Fpayload, whowhere)
		if err != nil {
			return pcargo, fmt.Errorf("cmd/tools/fetchData/1:\n%s\n====\n%s", err.Error(), catchErr)
		}
		if catchErr.Len() != 0 {
			return pcargo, fmt.Errorf("cmd/tools/fetchData/2:\n%s", catchErr)
		}
		dec := gob.NewDecoder(catchOut)
		pcargo = &qclient.Cargo{}
		for {
			if err := dec.Decode(pcargo); err == io.EOF {
				break
			} else if err != nil {
				return pcargo, fmt.Errorf("cmd/tools/fetchData/3:\n%s", err.Error())
			}
		}
	}
	return
}

func addData(ppayload *qclient.Payload, pcargo *qclient.Cargo, withcontent bool, batchid string) {
	query := ppayload.Query.Copy()
	psources := query.Run()
	paths := make([]string, len(psources))
	for i, ps := range psources {
		paths[i] = ps.String()
	}
	bodies := make([][]byte, 0)
	var errs error
	if withcontent {
		bodies, _, errs = qsource.FetchList(query.Release, paths)
	}
	buffer := new(bytes.Buffer)
	transports := make([]qclient.Transport, len(paths))

	for i, qpath := range paths {
		locfile := qclient.LocalFile{}
		pmeta, err := qmeta.Meta{}.New(query.Release, qpath)
		digest := "?"
		if err == nil {
			digest = pmeta.Digest
		}
		psource := psources[i]
		locfile.Release = query.Release
		locfile.QPath = qpath
		locfile.Project = psource.Project().String()
		locfile.Digest = digest
		locfile.Cu = pmeta.Cu
		locfile.Mu = pmeta.Mu
		locfile.Ct = pmeta.Ct
		locfile.Mt = pmeta.Mt
		transports[i].LocFile = locfile
		if withcontent && bodies[i] != nil {
			transports[i].Body = bodies[i]
		}
		if batchid != "" && strings.HasPrefix(batchid, "m:") {
			psource.ToMumps(batchid[2:], buffer)
			if !strings.HasSuffix(psource.String(), ".m") {
				qsource.Mend(batchid[2:], buffer)
			}
		}
		if batchid != "" && strings.HasPrefix(batchid, "r:") {
			err := psource.Resolve(batchid, nil, nil, buffer)
			pcargo.Error = err
		}

	}
	if batchid != "" {
		pcargo.Buffer = *buffer
	}
	pcargo.Transports = transports
	if withcontent {
		switch err := errs.(type) {
		case qerror.ErrorSlice:
			if len(err) == 0 {
				pcargo.Error = nil
			} else {
				pcargo.Error = err
			}
		default:
			if err == nil {
				pcargo.Error = nil
			} else {
				pcargo.Error = err
			}
		}
	}
}

func installData(ppayload *qclient.Payload, pcargo *qclient.Cargo, withcontent bool, batchid string) {
	query := ppayload.Query.Copy()
	psources := query.Run()
	if batchid == "" {
		batchid = "install"
	}
	errs := qinstall.Install(batchid, psources, true)
	switch err := errs.(type) {
	case qerror.ErrorSlice:
		if len(err) == 0 {
			pcargo.Error = nil
		} else {
			pcargo.Error = err
		}
	default:
		if err == nil {
			pcargo.Error = nil
		} else {
			pcargo.Error = err
		}
	}
}

func delData(ppayload *qclient.Payload, pcargo *qclient.Cargo) (errs error) {
	query := ppayload.Query.Copy()
	psources := query.Run()
	paths := make([]string, len(psources))
	transports := make([]qclient.Transport, len(paths))
	for i, ps := range psources {
		paths[i] = ps.String()
	}

	for i, qpath := range paths {
		locfile := qclient.LocalFile{}
		pmeta, err := qmeta.Meta{}.New(query.Release, qpath)
		digest := "?"
		if err == nil {
			digest = pmeta.Digest
		}
		psource := psources[i]
		locfile.Release = query.Release
		locfile.QPath = qpath
		locfile.Project = psource.Project().String()
		locfile.Digest = digest
		locfile.Cu = pmeta.Cu
		locfile.Mu = pmeta.Mu
		locfile.Ct = pmeta.Ct
		locfile.Mt = pmeta.Mt
		transports[i].LocFile = locfile
	}
	pcargo.Transports = transports

	r := query.Release

	errs = qsource.WasteList(r, paths)
	return errs
}

func listTransport(pcargo *qclient.Cargo) []lister {
	result := make([]lister, len(pcargo.Transports))
	if pcargo != nil && len(pcargo.Transports) != 0 {
		for i, transport := range Fcargo.Transports {
			locfil := transport.LocFile
			result[i] = lister{
				Release: locfil.Release,
				QPath:   locfil.QPath,
				Project: locfil.Project,
				Cu:      locfil.Cu,
				Mu:      locfil.Mu,
				Ct:      locfil.Ct,
				Mt:      locfil.Mt,
			}
		}
	}
	return result
}

func storeTransport() ([]storer, []error) {
	result := make([]storer, len(Fcargo.Transports))
	errlist := make([]error, 0)
	if Fcargo == nil || len(Fcargo.Transports) == 0 {
		return result, errlist
	}
	dirs := make(map[string][]int)
	idirs := make([]string, 0)
	for i, transport := range Fcargo.Transports {
		locfil := transport.LocFile

		dir := Fcwd
		if Fauto {
			Ftree = true
		}
		if Fauto && strings.ContainsRune(QtechType, 'W') {
			dir = qregistry.Registry["qtechng-workstation-basedir"]
			if dir == "" {
				dir = Fcwd
			}
		}
		qpath := locfil.QPath
		qdir, qbase := qutil.QPartition(qpath)
		tdir := Fcwd
		if Ftree {
			parts := strings.SplitN(qdir, "/", -1)
			parts[0] = tdir
			tdir = filepath.Join(parts...)
		}
		place := filepath.Join(tdir, qbase)
		locfil.Place = place
		Fcargo.Transports[i].LocFile = locfil
		dir = path.Dir(place)
		islice, ok := dirs[dir]
		if !ok {
			islice = make([]int, 0)
			idirs = append(idirs, dir)
		}
		islice = append(islice, i)
		dirs[dir] = islice
	}
	fn := func(n int) (interface{}, error) {
		errlist := make([]error, 0)
		dir := idirs[n]
		islice := dirs[dir]
		qfs.Mkdir(dir, "process")
		if !qfs.IsDir(dir) {
			err := qerror.QError{
				Ref: []string{"co.dir"},
				Msg: []string{"Cannot create `" + dir + "`"},
			}
			return nil, &err
		}
		oklocfils := make([]qclient.LocalFile, 0)
		for _, i := range islice {
			t := Fcargo.Transports[i]
			place := t.LocFile.Place
			body := t.Body
			e := qfs.Store(place, body, "")

			if e != nil {
				err := &qerror.QError{
					Ref:  []string{"co.store"},
					Type: "Error",
					File: place,
					Msg:  []string{"Cannot store file: `" + place + "`"},
				}
				errlist = append(errlist, err)
				continue
			}
			mt, e := qfs.GetMTime(place)
			if e == nil {
				touch := mt.Format(time.RFC3339)
				t.LocFile.Time = touch
			}
			oklocfils = append(oklocfils, t.LocFile)
		}
		d := new(qclient.Dir)
		d.Dir = dir
		d.Add(oklocfils...)
		if len(errlist) == 0 {
			return oklocfils, nil
		}
		return oklocfils, qerror.ErrorSlice(errlist)
	}
	resultlist, errorlist := qparallel.NMap(len(idirs), -1, fn)
	errs := Fcargo.Error

	if errs != nil {
		errlist = append(errlist, errs)
	}
	if len(errorlist) != 0 {
		errlist = append(errlist, errorlist...)
	}

	i := -1
	for _, locfils := range resultlist {
		for _, locfil := range locfils.([]qclient.LocalFile) {
			i++
			result[i] = storer{
				Release: locfil.Release,
				Qpath:   locfil.QPath,
				Project: locfil.Project,
				Time:    locfil.Time,
				Digest:  locfil.Digest,
				Cu:      locfil.Cu,
				Mu:      locfil.Mu,
				Ct:      locfil.Ct,
				Mt:      locfil.Mt,
				Place:   locfil.Place,
			}
		}
	}
	return result, errlist
}
