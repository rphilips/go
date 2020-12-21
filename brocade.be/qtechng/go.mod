module brocade.be/qtechng

go 1.15

replace (
	brocade.be/base/fnmatch => ../base/fnmatch
	brocade.be/base/fs => ../base/fs
	brocade.be/base/mumps => ../base/mumps
	brocade.be/base/parallel => ../base/parallel
	brocade.be/base/pattern => ../base/pattern
	brocade.be/base/python => ../base/python
	brocade.be/base/registry => ../base/registry
	brocade.be/base/ssh => ../base/ssh
	brocade.be/clipboard => ../clipboard
	brocade.be/qtechng/cli/cmd => ./cli/cmd
	brocade.be/qtechng/lib/client => ./lib/client
	brocade.be/qtechng/lib/error => ./lib/error
	brocade.be/qtechng/lib/file/bfile => ./lib/file/bfile
	brocade.be/qtechng/lib/file/dfile => ./lib/file/dfile
	brocade.be/qtechng/lib/file/ifile => ./lib/file/ifile
	brocade.be/qtechng/lib/file/lfile => ./lib/file/lfile
	brocade.be/qtechng/lib/file/mfile => ./lib/file/mfile
	brocade.be/qtechng/lib/file/objfile => ./lib/file/objfile
	brocade.be/qtechng/lib/file/ofile => ./lib/file/ofile
	brocade.be/qtechng/lib/file/xfile => ./lib/file/xfile
	brocade.be/qtechng/lib/meta => ./lib/meta
	brocade.be/qtechng/lib/object => ./lib/object
	brocade.be/qtechng/lib/project => ./lib/project
	brocade.be/qtechng/lib/server => ./lib/server
	brocade.be/qtechng/lib/source => ./lib/source
	brocade.be/qtechng/lib/util => ./lib/util
	brocade.be/qtechng/lib/vfs => ./lib/vfs
)

require (
	brocade.be/base/fnmatch v0.0.0
	brocade.be/base/fs v0.0.0
	brocade.be/base/mumps v0.0.0
	brocade.be/base/parallel v0.0.0-00010101000000-000000000000
	brocade.be/base/pattern v0.0.0 // indirect
	brocade.be/base/python v0.0.0
	brocade.be/base/registry v0.0.0
	brocade.be/base/ssh v0.0.0
	brocade.be/clipboard v0.0.0
	github.com/atotto/clipboard v0.1.2 // indirect
	github.com/benhoyt/goawk v1.6.1
	github.com/natefinch/atomic v0.0.0-20200526193002-18c0533a5b09 // indirect
	github.com/rwtodd/Go.Sed v0.0.0-20190103233418-906bc69c9394
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/afero v1.5.1
	github.com/spf13/cobra v1.1.1
	github.com/spyzhov/ajson v0.4.2
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/zloylos/grsync v0.0.0-20201103111841-a3c772f11d26
)
