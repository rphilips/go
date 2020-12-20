package ofile

import (
	qbfile "brocade.be/qtechng/lib/file/bfile"
	qdfile "brocade.be/qtechng/lib/file/dfile"
	qifile "brocade.be/qtechng/lib/file/ifile"
	qlfile "brocade.be/qtechng/lib/file/lfile"
	qxfile "brocade.be/qtechng/lib/file/xfile"
)

//DFile alias
type DFile = qdfile.DFile

//IFile alias
type IFile = qifile.IFile

//LFile alias
type LFile = qlfile.LFile

//BFile alias
type BFile = qbfile.BFile

//XFile alias
type XFile = qxfile.XFile

//Macro alias
type Macro = qdfile.Macro

//Include alias
type Include = qifile.Include

//Lgcode alias
type Lgcode = qlfile.Lgcode

//Widget alias
type Widget = qxfile.Widget
