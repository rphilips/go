package source

import (
	"bytes"

	qmumps "brocade.be/base/mumps"
	qbfile "brocade.be/qtechng/lib/file/bfile"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// BrobsListToMumps bereidt een verzameling van Brobs
func BrobsListToMumps(batchid string, brobs []*qbfile.Brob, buf *bytes.Buffer) {
	for _, pbrob := range brobs {
		mumps := pbrob.Mumps(batchid)
		qmumps.Println(buf, mumps)

	}
}

// BFileToMumps bereidt een verzameling van Brobs
func (bfile *Source) BFileToMumps(batchid string, buf *bytes.Buffer) error {

	content, err := bfile.Fetch()
	if err != nil {
		return err
	}
	content = qutil.Decomment(content).Bytes()
	content = qutil.About(content)
	bf := new(qofile.BFile)
	bf.SetEditFile(bfile.String())
	bf.SetRelease(bfile.Release().String())
	err = qobject.Loads(bf, content, true)
	if err != nil {
		return err
	}

	objectlist := bf.Objects()
	textmap := make(map[string]string)
	env := bfile.Env()

	notreplace := bfile.NotReplace()
	objectmap := make(map[string]qobject.Object)
	bufmac := new(bytes.Buffer)
	_, err = ResolveText(env, content, "rilm", notreplace, objectmap, textmap, bufmac, "", bfile.String())
	if err != nil {
		return err
	}

	content = bufmac.Bytes()

	err = qobject.Loads(bf, content, false)
	if err != nil {
		return err
	}
	brobs := make([]*qbfile.Brob, len(objectlist))
	for i, obj := range objectlist {
		brobs[i] = obj.(*qbfile.Brob)
	}
	BrobsListToMumps(batchid, brobs, buf)
	return err
}
