package server

import (
	"archive/tar"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	qvfs "brocade.be/qtechng/lib/vfs"
	_ "modernc.org/sqlite"
)

/// Backup

// TarBackup makes a backup of the sources and the meta information
func (release Release) SqliteBackup(sqlitefile string) error {
	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err = db.Exec(`
		CREATE TABLE sqlar (
			name TEXT PRIMARY KEY,
			mode INT,
  			mtime INT,
  			sz INT,
  			data BLOB
		);`); err != nil {
		return err
	}

	if _, err = db.Exec(`
		CREATE TABLE sources (
			name TEXT PRIMARY KEY,
			qpath TEXT,
			cu TEXT,
			ct TEXT,
			mu TEXT,
			mt TEXT,
			fu TEXT,
			ft TEXT
		);`); err != nil {
		return err
	}

	if _, err = db.Exec(`
		CREATE TABLE info (
			label TEXT PRIMARY KEY,
			value TEXT
		);`); err != nil {
		return err
	}

	stmt1, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, sz, data) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert1: %v", err)
	}
	defer stmt1.Close()

	stmt2, err := db.Prepare("INSERT INTO sources (name, qpath, cu, ct, mu, mt, fu, ft) Values($1,$2,$3,$4,$5,$6,$7,$8)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt2.Close()

	stmt3, err := db.Prepare("INSERT INTO info (label, value) Values($1,$2)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt3.Close()

	rfs := release.FS("/")

	stmt3.Exec("system-name", qregistry.Registry["system-name"])
	stmt3.Exec("dns-name", qregistry.Registry["dns-name"])
	h := time.Now()
	stmt3.Exec("begin-time", h.Format(time.RFC3339))

	for _, dir := range []string{"/source/data"} {
		source, err := rfs.RealPath(dir)
		if err != nil {
			return err
		}

		fn := func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := qutil.AbsPath(p, source)
			qpath, _ := filepath.Rel(source, name)
			qpath = qutil.Canon(qpath)

			// archive
			data, err := qfs.Fetch(name)
			if err != nil {
				return fmt.Errorf("cannot get content of `%s`: %v", name, err)
			}
			mt, err := qfs.GetMTime(name)
			if err != nil {
				return fmt.Errorf("cannot get mtime of `%s`: %v", name, err)
			}
			sz, err := qfs.GetSize(name)
			if err != nil {
				return fmt.Errorf("cannot get size of `%s`: %v", name, err)
			}
			mode, err := qfs.GetPerm(name)
			if err != nil {
				return fmt.Errorf("cannot get access permissions of `%s`: %v", name, err)
			}
			mtime := mt.Unix()
			_, err = stmt1.Exec(name, uint32(mode), mtime, sz, data)
			if err != nil {
				return fmt.Errorf("cannot exec: %v", err)
			}

			// meta
			fs, place := release.MetaPlace(qpath)
			meta, _ := fs.ReadFile(place)
			m := make(map[string]string)
			if len(meta) != 0 {
				json.Unmarshal(meta, &m)
			}
			_, err = stmt2.Exec(name, qpath, m["cu"], m["ct"], m["mu"], m["mt"], m["fu"], m["ft"])
			if err != nil {
				return fmt.Errorf("cannot exec: %v", err)
			}
			return nil
		}

		err = filepath.Walk(source, fn)
		if err != nil {
			return err
		}

	}
	h = time.Now()
	stmt3.Exec("end-time", h.Format(time.RFC3339))
	return nil
}

// TarBackup makes a backup of the sources and the meta information
func (release Release) TarBackup(tarfile string) error {
	ftar, err := os.Create(tarfile)

	if err != nil {
		return err
	}
	defer ftar.Close()
	errs := make([]error, 0)

	tw := tar.NewWriter(ftar)
	defer tw.Close()

	fs := release.FS("/")
	for _, dir := range []string{"/source/data", "/meta"} {
		source, err := fs.RealPath(dir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		info, err := os.Stat(source)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var baseDir string
		if info.IsDir() {
			baseDir = dir[1:]
		}
		err = filepath.Walk(source,
			func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				header, err := tar.FileInfoHeader(info, info.Name())
				if err != nil {
					return err
				}

				if baseDir != "" {
					header.Name = filepath.Join(baseDir, filepath.ToSlash(strings.TrimPrefix(p, source)))
				}

				if err := tw.WriteHeader(header); err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				file, err := os.Open(p)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				return err
			})
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	tw.Flush()
	tw.Close()
	ftar.Close()
	if len(errs) == 0 {
		return nil
	}
	return qerror.ErrorSlice(errs)
}

// Restore restores a backup and the meta information
func (release Release) Restore(tarfile string, init bool) (previous string, err error) {

	reader, err := os.Open(tarfile)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// create backup
	h := time.Now()
	t := h.Format(time.RFC3339)[:19]
	t = strings.ReplaceAll(t, ":", "")
	t = strings.ReplaceAll(t, "-", "")
	r := release.String()
	previous = filepath.Join(filepath.Dir(tarfile), "previous-brocade-"+r+"-"+t+".tar")
	err = release.TarBackup(previous)
	if err != nil {
		return "", err
	}

	// initialises if necessary
	fs := release.FS("/")
	if init {
		fs := release.FS("/")
		for _, dir := range []string{"/source/data", "/meta"} {
			source, _ := fs.RealPath(dir)
			qfs.Rmpath(source)
			if qfs.Exists(source) {
				return previous, fmt.Errorf("cannot remove `%s`", source)
			}
			qfs.Mkdir(source, "qtech")
			if !qfs.IsDir(source) {
				return previous, fmt.Errorf("cannot create directory `%s`", source)
			}
		}
	}

	// restore files

	tarReader := tar.NewReader(reader)
	target, _ := fs.RealPath("/")
	if !qfs.IsDir(target) {
		return previous, fmt.Errorf("cannot write to directory `%s`", target)
	}

	dirs := make(map[string]bool)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return previous, fmt.Errorf("cannot read tarfile `%s`", tarfile)
		}
		info := header.FileInfo()
		if info.IsDir() {
			continue
		}
		fname := header.Name
		ismeta := false
		var fs *qvfs.QFs
		path := ""
		var body []byte
		if strings.HasPrefix(fname, "meta") {
			ismeta = true
			buffer := bytes.NewBuffer([]byte{})
			_, err := io.Copy(buffer, tarReader)
			if err != nil {
				return previous, fmt.Errorf("cannot read file `%s`", fname)
			}
			m := make(map[string]string)
			body = buffer.Bytes()
			err = json.Unmarshal(body, &m)
			if err != nil {
				return previous, fmt.Errorf("file `%s` is not a meta file", fname)
			}
			source := m["source"]
			if source == "" {
				return previous, fmt.Errorf("file `%s` is not a suitable meta file: missing source", fname)
			}
			fs, path = release.MetaPlace(source)
		}
		if path == "" {
			fname = filepath.ToSlash(fname)
			if strings.HasPrefix(fname, "source/") {
				fname = strings.SplitN(fname, "source/", 2)[1]
			}
			if strings.HasPrefix(fname, "data/") {
				fname = strings.SplitN(fname, "data/", 2)[1]
			}
			fname = "/" + strings.Trim(fname, "/")
			fs, path = release.SourcePlace(fname)
		}
		rpath, _ := fs.RealPath(path)
		rdir := filepath.Dir(rpath)
		if !dirs[rdir] {
			if err = qfs.MkdirAll(rdir, "qtech"); err != nil {
				return previous, fmt.Errorf("cannot make directory `%s`", rdir)
			}
			dirs[rdir] = true
		}
		if ismeta {
			err := qfs.Store(rpath, body, "qtech")
			if err != nil {
				return previous, fmt.Errorf("cannot store to `%s`", rpath)
			}
			continue
		}
		file, err := os.OpenFile(rpath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return previous, fmt.Errorf("cannot make file `%s` with mode `%s` (error `%s`)", rpath, info.Mode(), err)
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return previous, fmt.Errorf("cannot write to file `%s` with mode `%s`", path, info.Mode())
		}
		file.Close()
	}
	return previous, nil

}
