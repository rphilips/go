package server

import (
	"strings"

	qutil "brocade.be/qtechng/lib/util"
	qvfs "brocade.be/qtechng/lib/vfs"
)

// sources

func (release *Release) SourcePlace(qpath string) (*qvfs.QFs, string) {
	fs := release.FS()
	return &fs, qpath
}

func (release *Release) UniquePlace(qpath string) (*qvfs.QFs, string) {
	_, base := qutil.QPartition(qpath)
	digest := qutil.Digest([]byte(base))
	ndigest := qutil.Digest([]byte(qpath))
	fs := release.FS("/unique")
	fname := "/" + digest[:2] + "/" + digest[2:] + "/" + ndigest
	return &fs, fname
}

func (release *Release) MetaPlace(qpath string) (*qvfs.QFs, string) {
	fs := release.FS("/meta")
	digest := qutil.Digest([]byte(qpath))
	place := "/" + digest[0:2] + "/" + digest[2:] + ".json"
	return &fs, place
}

// Objects

func (release *Release) ObjectPlace(objname string) (*qvfs.QFs, string) {
	if objname == "" {
		fs := release.FS("object", "")
		return &fs, ""
	}
	if !strings.ContainsRune(objname, '_') {
		fs := release.FS("object", objname)
		return &fs, ""
	}
	ty := strings.SplitN(objname, "_", 2)[0]
	objname, _ = qutil.DeNEDFU(objname)
	fs := release.FS("object", ty)
	h := qutil.Digest([]byte(objname))
	dirname := "/" + h[0:2] + "/" + h[2:]
	return &fs, dirname + "/obj.json"
}

func (release *Release) ObjectDepPlace(objname string, fname string) (*qvfs.QFs, string) {
	fs, place := release.ObjectPlace(objname)
	digest := qutil.Digest([]byte(fname))
	place = strings.TrimSuffix(place, "obj.json") + digest[:2] + "/" + digest[2:] + ".dep"
	return fs, place
}
