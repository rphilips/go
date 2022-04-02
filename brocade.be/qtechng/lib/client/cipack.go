package client

import (
	"os"
	"path/filepath"

	qfnmatch "brocade.be/base/fnmatch"
	qfs "brocade.be/base/fs"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

// CiPack basic data structure for a file to send out over the wire
type CiPack struct {
	Release string
	QPath   string
	Digest  string
	Force   bool
	Body    []byte
}

// Pack build a slice of Packs to transfer to B-machine
func Pack(cwd string, files []string, release string, qpattern string, force bool) (result []CiPack, err error) {
	find := false
	if len(files) == 0 {
		find = true
		files, err = qfs.Find(cwd, nil, true, true, false)
		if err != nil {
			err := &qerror.QError{
				Ref:  []string{"cipack.pack.find"},
				Type: "Error",
				Msg:  []string{"Cannot find files: " + err.Error()},
			}
			return nil, err
		}
	}
	done := make(map[string]bool)

	result = make([]CiPack, 0)
	errlist := make([]error, 0)

	for _, file := range files {
		if done[file] {
			continue
		}
		done[file] = true
		place := qutil.AbsPath(file, cwd)
		dir := filepath.Dir(place)
		d := new(Dir)
		d.Dir = dir
		base := filepath.Base(place)
		plocfil := d.Get(base)
		if plocfil == nil {
			if !find {
				err := &qerror.QError{
					Ref:  []string{"cipack.pack.get"},
					Type: "Error",
					Msg:  []string{"`" + file + "` does not exists in QtechNG"},
				}
				errlist = append(errlist, err)
			}
			continue
		}
		if release != "" {
			rok := qfnmatch.Match(release, plocfil.Release)
			if !rok && !find {
				err := &qerror.QError{
					Ref:  []string{"cipack.pack.version"},
					Type: "Error",
					Msg:  []string{"`" + file + "` does not match version"},
				}
				errlist = append(errlist, err)
			}
			if !rok {
				continue
			}
		}
		if qpattern != "" {
			qok := qfnmatch.Match(qpattern, plocfil.QPath)
			if !qok && !find {
				err := &qerror.QError{
					Ref:  []string{"cipack.pack.qpath"},
					Type: "Error",
					Msg:  []string{"`" + file + "` does not match pattern"},
				}
				errlist = append(errlist, err)
			}
			if !qok {
				continue
			}
		}

		body, err := os.ReadFile(file)

		if err != nil {
			err := &qerror.QError{
				Ref:  []string{"cipack.pack.read"},
				Type: "Error",
				Msg:  []string{"`" + file + "`: " + err.Error()},
			}
			errlist = append(errlist, err)
			continue
		}

		cipack := CiPack{
			Release: plocfil.Release,
			QPath:   plocfil.QPath,
			Digest:  plocfil.Digest,
			Force:   force,
			Body:    body,
		}
		result = append(result, cipack)
	}
	if len(result) == 0 {
		result = nil
	}
	if len(errlist) == 0 {
		return result, nil
	}
	return result, qerror.ErrorSlice(errlist)
}
