package cmd

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

type lister struct {
	Release string `json:"version"`
	QPath   string `json:"qpath"`
	DevPath string `json:"devpath"`
	Project string `json:"project"`
	Path    string `json:"file"`
	URL     string `json:"fileurl"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
}

type linter struct {
	Release string `json:"version"`
	QPath   string `json:"qpath"`
	Project string `json:"project"`
	Path    string `json:"file"`
	URL     string `json:"fileurl"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
	Info    string `json:"info"`
}

type projlister struct {
	Release string `json:"version"`
	Project string `json:"project"`
	Sort    string `json:"sort"`
}
type storer struct {
	Release string `json:"version"`
	QPath   string `json:"qpath"`
	Project string `json:"project"`
	URL     string `json:"fileurl"`
	Changed bool   `json:"changed"`
	Time    string `json:"time"`
	Digest  string `json:"digest"`
	Cu      string `json:"cu"`
	Mu      string `json:"mu"`
	Ct      string `json:"ct"`
	Mt      string `json:"mt"`
	Place   string `json:"file"`
}

func buildSQuery(args []string, filesinproject bool, qdirs []string, mumps bool) qsource.SQuery {

	if len(args) != 0 {
		if len(Fqpattern) == 0 {
			Fqpattern = args
		} else {
			for _, arg := range args {
				ok := len(Fqpattern) == 0
				for _, p := range Fqpattern {
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
					Fqpattern = append(Fqpattern, arg)
				}
			}
		}
	}

	return qsource.SQuery{
		Release:        Fversion,
		CmpRelease:     Fcmpversion,
		QDirs:          qdirs,
		Patterns:       Fqpattern,
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
		SmartCase:      Fsmartcase,
		Contains:       Fneedle,
		Mumps:          mumps,
	}

}

func fetchData(args []string, filesinproject bool, qdirs []string, mumps bool) (pcargo *qclient.Cargo, err error) {

	Fpayload = &qclient.Payload{
		ID:     "Once",
		UID:    FUID,
		CMD:    "qtechng",
		Origin: QtechType,
		Args:   os.Args[1:],
		Query:  buildSQuery(args, filesinproject, qdirs, mumps),
	}

	pcargo = new(qclient.Cargo)
	if strings.ContainsAny(QtechType, "PB") {
		return
	}
	whowhere := qregistry.Registry["qtechng-user"] + "@" + qregistry.Registry["qtechng-server"]
	catchOut, catchErr, err := qssh.SSHcmd(Fpayload, whowhere)
	if err != nil {
		return pcargo, fmt.Errorf("cmd/tools/fetchData/1:\n%s\n====\n%s", err.Error(), catchErr)
	}
	if catchErr.Len() != 0 {
		return pcargo, fmt.Errorf("cmd/tools/fetchData/2:\n%s", catchErr)
	}
	dec := gob.NewDecoder(catchOut)
	err = dec.Decode(pcargo)
	if err == io.EOF {
		err = nil
	}
	if err != nil {
		return pcargo, fmt.Errorf("cmd/tools/fetchData/3:\n%s", err.Error())
	}
	return
}

func fetchObjectData(args []string) (pcargo *qclient.Cargo, err error) {

	if len(args) != 0 {
		if len(Fobjpattern) == 0 {
			Fobjpattern = args
		} else {
			for _, arg := range args {
				ok := len(Fobjpattern) == 0
				for _, p := range Fobjpattern {
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
					Fobjpattern = append(Fobjpattern, arg)
				}
			}
		}
	}

	squery := qsource.SQuery{
		Release: Fversion,
		Objects: Fobjpattern,
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
		whowhere := qregistry.Registry["qtechng-user"] + "@" + qregistry.Registry["qtechng-server"]
		catchOut, catchErr, err := qssh.SSHcmd(Fpayload, whowhere)
		if err != nil {
			return pcargo, fmt.Errorf("cmd/tools/fetchObject/1:\n%s\n====\n%s", err.Error(), catchErr)
		}
		if catchErr.Len() != 0 {
			return pcargo, fmt.Errorf("cmd/tools/fetchObject/2:\n%s", catchErr)
		}
		dec := gob.NewDecoder(catchOut)
		pcargo = &qclient.Cargo{}
		for {
			if err := dec.Decode(pcargo); err == io.EOF {
				break
			} else if err != nil {
				return pcargo, fmt.Errorf("cmd/tools/fetchObject/3:\n%s", err.Error())
			}
		}
	}
	return
}

func addData(ppayload *qclient.Payload, pcargo *qclient.Cargo, withcontent bool, withlint bool, batchid string) {
	query := ppayload.Query.Copy()

	psources := query.Run()
	paths := make([]string, len(psources))
	for i, ps := range psources {
		paths[i] = ps.String()
	}
	bodies := make([][]byte, 0)
	infos := make([][]byte, 0)
	var errs error

	if withcontent {
		bodies, _, errs = qsource.FetchList(query.Release, paths)
	}
	if withlint {
		var einfos []error
		einfos, _, errs = qsource.LintList(query.Release, paths, strings.HasPrefix(batchid, "w:"))
		for _, einfo := range einfos {
			if einfo == nil {
				infos = append(infos, []byte("OK"))
			} else {
				infos = append(infos, []byte(einfo.Error()))
			}
		}
	}
	buffer := new(bytes.Buffer)
	transports := make([]qclient.Transport, len(paths))

	sorts := make(map[string]string)

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
		if strings.Contains(qregistry.Registry["qtechng-type"], "B") {
			locfile.DevPath = psource.Path()
		}
		locfile.Project = psource.Project().String()
		locfile.Digest = digest
		locfile.Cu = pmeta.Cu
		locfile.Mu = pmeta.Mu
		locfile.Ct = pmeta.Ct
		locfile.Mt = pmeta.Mt
		proj := locfile.Project
		if sorts[proj] == "" {
			sorts[proj] = psource.Project().Orden()
		}
		locfile.Sort = sorts[proj]
		transports[i].LocFile = locfile
		if withcontent && bodies[i] != nil {
			transports[i].Body = bodies[i]
		}
		if withlint && infos[i] != nil {
			transports[i].Info = infos[i]
		}
		if batchid != "" && strings.HasPrefix(batchid, "m:") {
			psource.ToMumps(batchid[2:], buffer)
			if !strings.HasSuffix(psource.String(), ".m") {
				qsource.Mend(batchid[2:], buffer)
			}
		}
		if batchid != "" && strings.HasPrefix(batchid, "r:") {
			err := psource.Resolve(batchid, nil, nil, buffer, false)
			pcargo.AddError(err)
		}

	}
	pcargo.Data = make([]byte, 0)
	if batchid != "" {
		pcargo.Data = buffer.Bytes()
	}
	pcargo.Transports = transports
	if withcontent || withlint {
		pcargo.AddError(errs)
	}
}

func addObjectData(ppayload *qclient.Payload, pcargo *qclient.Cargo, batchid string) {
	query := ppayload.Query.Copy()
	pubermap := query.RunObject()
	b, _ := json.Marshal(pubermap)
	buffer := bytes.NewBuffer(b)
	pcargo.Data = buffer.Bytes()
}

func rebuildData(ppayload *qclient.Payload, pcargo *qclient.Cargo, withcontent bool, batchid string) {
	query := ppayload.Query.Copy()
	psources := query.Run()
	if len(psources) == 0 {
		return
	}
	version := psources[0].Release().String()
	if batchid == "" {
		batchid = "rebuild"
	}
	qpaths := make([]string, len(psources))
	for i, psource := range psources {
		qpaths[i] = psource.String()
	}
	errs := qsource.Rebuild(batchid, version, qpaths)
	pcargo.AddError(errs)
}

// func delData(ppayload *qclient.Payload, pcargo *qclient.Cargo) (errs error) {
// 	query := ppayload.Query.Copy()
// 	psources := query.Run()
// 	paths := make([]string, len(psources))
// 	transports := make([]qclient.Transport, len(paths))
// 	for i, ps := range psources {
// 		paths[i] = ps.String()
// 	}

// 	for i, qpath := range paths {
// 		locfile := qclient.LocalFile{}
// 		pmeta, err := qmeta.Meta{}.New(query.Release, qpath)
// 		digest := "?"
// 		if err == nil {
// 			digest = pmeta.Digest
// 		}
// 		psource := psources[i]
// 		locfile.Release = query.Release
// 		locfile.QPath = qpath
// 		locfile.Project = psource.Project().String()
// 		locfile.Digest = digest
// 		locfile.Cu = pmeta.Cu
// 		locfile.Mu = pmeta.Mu
// 		locfile.Ct = pmeta.Ct
// 		locfile.Mt = pmeta.Mt
// 		transports[i].LocFile = locfile
// 	}
// 	pcargo.Transports = transports

// 	r := query.Release

// 	pcargo.Error = qsource.WasteList(r, paths)
// 	return pcargo.Error
// }

func delData(squery qsource.SQuery, number int) (qpaths []string, errs error) {
	query := squery.Copy()
	psources := query.Run()
	if len(psources) == 0 {
		return nil, nil
	}
	qpaths = make([]string, len(psources))
	for i, ps := range psources {
		qpaths[i] = ps.String()
	}

	if number >= 0 && number != len(qpaths) {
		return qpaths, fmt.Errorf("there are %d sources to be deleted. Given number is %d. Use the `--number=%d` flag", len(qpaths), number, len(qpaths))
	}

	r := query.Release
	errs = qsource.WasteList(r, qpaths)
	if errs != nil {
		return nil, errs
	}
	return qpaths, nil
}

func renameData(squery qsource.SQuery, number int, replace string, with string, regexp bool, overwrite bool) (qrenames map[string]string, errs []error) {
	query := squery.Copy()
	psources := query.Run()
	if len(psources) == 0 {
		return nil, nil
	}
	qpaths := make([]string, len(psources))
	for i, ps := range psources {
		qpaths[i] = ps.String()
	}
	qrenames, errs = calcRenames(qpaths, replace, with, regexp)
	if errs != nil {
		return qrenames, errs
	}

	if number >= 0 && number != len(qpaths) {
		errs = append(errs, fmt.Errorf("there are %d sources to be renamed. Given number is %d. Use the `--number=%d` flag", len(qpaths), number, len(qpaths)))
		return qrenames, errs
	}

	if len(errs) == 0 {
		e := qsource.Rename(query.Release, qrenames, overwrite)
		if e != nil {
			errs = append(errs, e)
		}
	}

	return qrenames, errs
}

func calcRenames(qpaths []string, replace string, with string, regxp bool) (qrenames map[string]string, errs []error) {
	if len(qpaths) == 0 {
		return nil, nil
	}
	doubles := make(map[string]bool)
	qrenames = make(map[string]string)
	var rex *regexp.Regexp
	if replace != "" && regxp {
		rex, _ = regexp.Compile(replace)
	}
	for _, qpath := range qpaths {
		_, ok := qrenames[qpath]
		if ok {
			continue
		}
		ren := ""
		if replace == "" {
			ren = with
		}
		if ren == "" && !regxp {
			ren = strings.ReplaceAll(qpath, replace, with)
		}
		if ren == "" && regxp {
			ren = rex.ReplaceAllString(qpath, with)
		}
		if ren == "" {
			errs = append(errs, fmt.Errorf("`%s` is renamed to empty", qpath))
			continue
		}
		if ren == "/" {
			errs = append(errs, fmt.Errorf("`%s` is renamed to `/`", qpath))
			continue
		}
		ren = qutil.Canon(ren)
		if doubles[ren] {
			errs = append(errs, fmt.Errorf("`%s` found more than once", ren))
			continue
		}
		if ren == qpath {
			continue
		}
		qrenames[qpath] = ren
		doubles[ren] = true
	}
	return
}

func listTransport(pcargo *qclient.Cargo) ([]string, []lister) {
	if pcargo == nil {
		return nil, nil
	}
	result := make([]lister, len(pcargo.Transports))
	qpaths := make([]string, len(pcargo.Transports))
	if len(pcargo.Transports) != 0 {
		for i, transport := range Fcargo.Transports {
			locfil := transport.LocFile
			qpaths[i] = locfil.QPath
			result[i] = lister{
				Release: locfil.Release,
				QPath:   locfil.QPath,
				DevPath: locfil.DevPath,
				Project: locfil.Project,
				Path:    locfil.Place,
				URL:     qutil.FileURL(locfil.Place, locfil.QPath, -1),
				Cu:      locfil.Cu,
				Mu:      locfil.Mu,
				Ct:      locfil.Ct,
				Mt:      locfil.Mt,
			}
		}
	}

	return qpaths, result
}

func lintTransport(pcargo *qclient.Cargo) ([]string, []linter) {
	if pcargo == nil {
		return nil, nil
	}
	result := make([]linter, len(pcargo.Transports))
	qpaths := make([]string, 0)
	if len(pcargo.Transports) != 0 {
		for i, transport := range Fcargo.Transports {
			locfil := transport.LocFile

			result[i] = linter{
				Release: locfil.Release,
				QPath:   locfil.QPath,
				Project: locfil.Project,
				Path:    locfil.Place,
				URL:     qutil.FileURL(locfil.Place, locfil.QPath, -1),
				Cu:      locfil.Cu,
				Mu:      locfil.Mu,
				Ct:      locfil.Ct,
				Mt:      locfil.Mt,
				Info:    string(transport.Info),
			}
			if info := strings.ToUpper(result[i].Info); info != "NOLINT" && info != "OK" && info != "" {
				qpaths = append(qpaths, locfil.QPath)
			}
		}
	}
	return qpaths, result
}

func listObjectTransport(pcargo *qclient.Cargo) []byte {
	return pcargo.Data
}

func storeTransport(dirname string, qdir string) ([]string, []storer, []error) {
	result := make([]storer, len(Fcargo.Transports))
	qpaths := make([]string, len(Fcargo.Transports))
	errlist := make([]error, 0)
	if Fcargo == nil || len(Fcargo.Transports) == 0 {
		return nil, result, errlist
	}
	dirs := make(map[string][]int)
	idirs := make([]string, 0)

	coredir := dirname

	if Fauto {
		Ftree = true
	}
	if Froot {
		Ftree = true
	}

	if Fauto && strings.ContainsRune(QtechType, 'W') {
		coredir = qregistry.Registry["qtechng-work-dir"]
		if coredir == "" {
			coredir = dirname
			Fclear = false
			Fauto = false
		}
	}
	root := ""
	if Froot {
		for _, transport := range Fcargo.Transports {
			locfil := transport.LocFile
			qpath := locfil.QPath
			dir, _ := qutil.QPartition(qpath)
			dir += "/"
			if root == "" {
				root = dir
				continue
			}
			if strings.HasPrefix(dir, root) {
				continue
			}
			for {
				if root == "/" {
					break
				}
				root = strings.TrimRight(root, "/")
				qroot, _ := qutil.QPartition(root)
				if len(qroot) < 2 {
					root = "/"
					break
				} else {
					root = qroot + "/"
				}
				if strings.HasPrefix(dir, root) {
					break
				}
			}
		}
		root = strings.TrimRight(root, "/")
	}

	for i, transport := range Fcargo.Transports {
		locfil := transport.LocFile
		qpath := locfil.QPath
		place := ""
		if Ftree {
			qp := qpath
			if root != "" {
				qp = strings.TrimPrefix(qpath, root)
			}
			parts := strings.SplitN(qp, "/", -1)
			parts[0] = coredir
			place = filepath.Join(parts...)
		} else {
			_, qbase := qutil.QPartition(qpath)
			place = filepath.Join(coredir, qbase)
		}
		locfil.Place = place
		Fcargo.Transports[i].LocFile = locfil
		dir := filepath.Dir(place)
		islice, ok := dirs[dir]
		if !ok {
			islice = make([]int, 0)
			idirs = append(idirs, dir)
		}
		islice = append(islice, i)
		dirs[dir] = islice
	}
	if Fclear && !Ftransported {
		for _, dir := range idirs {
			qfs.Rmpath((dir))
		}
	}

	dur, _ := time.ParseDuration("-1s")
	stamp := time.Now().Add(dur)

	fn := func(n int) (interface{}, error) {
		errlist := make([]error, 0)
		dir := idirs[n]
		islice := dirs[dir]
		e := qfs.Mkdir(dir, "process")
		if !qfs.IsDir(dir) {
			extra := ""
			if e != nil {
				extra = ": " + e.Error()
			}
			err := qerror.QError{
				Ref: []string{"co.dir"},
				Msg: []string{"Cannot create `" + dir + "`" + extra},
			}
			return nil, &err
		}
		oklocfils := make([]qclient.LocalFile, 0)
		for _, i := range islice {
			t := Fcargo.Transports[i]
			place := t.LocFile.Place
			body := t.Body
			e := qfs.Store(place, body, "qtech")

			if e != nil {
				err := &qerror.QError{
					Ref:  []string{"co.store"},
					Type: "Error",
					File: place,
					Msg:  []string{"Cannot store file: `" + place + "`" + ":" + e.Error()},
				}
				errlist = append(errlist, err)
				continue
			}
			os.Chtimes(place, stamp, stamp)
			mt, e := qfs.GetMTime(place)
			if e == nil {
				touch := mt.Format(time.RFC3339)
				t.LocFile.Time = touch
			}
			oklocfils = append(oklocfils, t.LocFile)
		}
		if !Fcopyonly {
			d := new(qclient.Dir)
			d.Dir = dir
			d.Add(oklocfils...)
		}
		if len(errlist) == 0 {
			return oklocfils, nil
		}
		return oklocfils, qerror.ErrorSlice(errlist)
	}
	resultlist, errorlist := qparallel.NMap(len(idirs), -1, fn)
	if len(errorlist) != 0 {
		for _, e := range errorlist {
			if e == nil {
				continue
			}
			errlist = append(errlist, e)
			Fcargo.AddError(e)
		}
	}

	i := -1
	for _, locfils := range resultlist {
		if locfils == nil {
			continue
		}
		for _, locfil := range locfils.([]qclient.LocalFile) {
			i++
			qpaths[i] = locfil.QPath
			result[i] = storer{
				Release: locfil.Release,
				QPath:   locfil.QPath,
				Project: locfil.Project,
				URL:     qutil.FileURL(locfil.Place, locfil.QPath, -1),
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
	return qpaths, result, errlist
}

func glob(cwd string, args []string, recurse bool, patterns []string, fils bool, dirs bool, onlyutf8 bool) (files []string, err error) {

	for _, arg := range args {
		arg = qutil.AbsPath(arg, cwd)
		if qfs.IsDir(arg) {
			paths, err := qfs.Find(arg, patterns, recurse, fils, dirs)
			if err != nil {
				return nil, err
			}
			files = append(files, paths...)
			continue
		}
		if !qfs.IsFile(arg) {
			return nil, fmt.Errorf("`%s` is not a file", arg)
		}
		if !fils {
			continue
		}
		if len(patterns) == 0 {
			files = append(files, arg)
			continue
		}
		ok := false
		basename := filepath.Base(arg)
		for _, pattern := range patterns {
			ok, _ = path.Match(pattern, basename)
			if ok {
				break
			}
		}
		if ok {
			files = append(files, arg)
		}
	}
	if !onlyutf8 {
		return
	}

	fn := func(n int) (interface{}, error) {
		fname := files[n]
		r := validutf8(fname)
		if r {
			return fname, nil
		}
		return "", nil
	}
	resultlist, _ := qparallel.NMap(len(files), -1, fn)
	result := make([]string, 0)

	for _, r := range resultlist {
		fname := r.(string)
		if fname != "" {
			result = append(result, fname)
		}
	}
	return result, nil
}

func validutf8(fname string) bool {
	f, err := os.Open(fname)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := bytes.NewBuffer(make([]byte, 8192))
	reader := bufio.NewReader(f)
	maxlen := 0
	for {
		text, err := reader.ReadString('\n')
		if err == io.EOF {
			buf.WriteString(text)
			break
		}
		if err != nil {
			return false
		}
		if strings.ContainsRune(text, 0) {
			return false
		}
		buf.WriteString(text)
		maxlen += len(text)
		if maxlen > 4096 {
			break
		}
	}
	return utf8.ValidString(buf.String())
}
