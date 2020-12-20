package source

import (
	"bytes"

	qmumps "brocade.be/base/mumps"
	qbfile "brocade.be/qtechng/lib/file/bfile"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
)

// BrobsListToMumps bereidt een verzameling van Brobs
func BrobsListToMumps(batchid string, brobs []*qbfile.Brob, buf *bytes.Buffer) {
	for _, pbrob := range brobs {
		mumps := pbrob.Mumps(batchid)
		qmumps.Println(buf, mumps)

	}
	return
}

// BFileToMumps bereidt een verzameling van Brobs
func (bfile *Source) BFileToMumps(batchid string, buf *bytes.Buffer) {
	content, err := bfile.Fetch()
	if err != nil {
		return
	}
	objfile := new(qofile.BFile)
	objfile.SetEditFile(bfile.String())
	objfile.SetRelease(bfile.Release().String())
	err = qobject.Loads(objfile, content)
	if err != nil {
		return
	}
	objectlist := objfile.Objects()
	brobs := make([]*qbfile.Brob, len(objectlist))
	for i, obj := range objectlist {
		brobs[i] = obj.(*qbfile.Brob)
	}
	BrobsListToMumps(batchid, brobs, buf)
	return
}
