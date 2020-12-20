module brocade.be/qtechng/lib/lib

go 1.15

replace brocade.be/base/fnmatch => ../../../brocade.be/base/fnmatch

replace brocade.be/base/fs => ../../../brocade.be/base/fs

replace brocade.be/base/mumps => ../../../brocade.be/base/mumps

replace brocade.be/base/registry => ../../../brocade.be/base/registry

replace brocade.be/base/parallel => ../../../brocade.be/base/parallel
replace brocade.be/base/pattern => ../../../brocade.be/base/pattern

require (
	brocade.be/base/fnmatch v0.0.0
	brocade.be/base/fs v0.0.0
	brocade.be/base/pattern v0.0.0
	brocade.be/base/mumps v0.0.0
	brocade.be/base/parallel v0.0.0-00010101000000-000000000000
	brocade.be/base/registry v0.0.0
	github.com/natefinch/atomic v0.0.0-20200526193002-18c0533a5b09 // indirect
	github.com/spf13/afero v1.5.1
	github.com/spyzhov/ajson v0.4.2
)
