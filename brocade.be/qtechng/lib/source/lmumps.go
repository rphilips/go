package source

import (
	"bytes"

	qmumps "brocade.be/base/mumps"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
)

// LgcodesSetToMumps bereidt een verzameling van Lgcodes
func LgcodesSetToMumps(batchid string, r string, lgcodes map[string]bool, buf *bytes.Buffer) {
	for l4, tf := range lgcodes {
		if !tf {
			continue
		}
		lgcode := &qofile.Lgcode{
			ID:      l4,
			Version: r,
		}
		err := qobject.Fetch(lgcode)
		if err != nil {
			continue
		}
		mumps := lgcode.Mumps(batchid)
		qmumps.Println(buf, mumps)

	}
}

// LgcodesListToMumps bereidt een verzameling van Lgcodes
func LgcodesListToMumps(batchid string, lgcodes []*qofile.Lgcode, buf *bytes.Buffer) {
	for _, plgcode := range lgcodes {
		mumps := plgcode.Mumps(batchid)
		qmumps.Println(buf, mumps)

	}
}

// LFileToMumps bereidt een verzameling van Lgcodes
func (lfile *Source) LFileToMumps(batchid string, buf *bytes.Buffer) {
	content, err := lfile.Fetch()
	if err != nil {
		return
	}
	objfile := new(qofile.LFile)
	objfile.SetEditFile(lfile.String())
	objfile.SetRelease(lfile.Release().String())
	err = qobject.Loads(objfile, content, true)
	if err != nil {
		return
	}
	objectlist := objfile.Objects()
	lgcodes := make([]*qofile.Lgcode, len(objectlist))
	for i, obj := range objectlist {
		lgcodes[i] = obj.(*qofile.Lgcode)
	}
	LgcodesListToMumps(batchid, lgcodes, buf)
}

// Mend ends a batch write to M
func Mend(batchid string, buf *bytes.Buffer) {
	m := qmumps.M{
		Subs:   []string{"batchid"},
		Value:  batchid,
		Action: "set",
	}
	mumps := qmumps.MUMPS{m}
	m = qmumps.M{
		Value:  "d %Run^bqtin(batchid)",
		Action: "exec",
	}
	mumps = append(mumps, m)
	qmumps.Println(buf, mumps)
}
