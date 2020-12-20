module brocade.be/qtechng/cli

go 1.15

replace (
	brocade.be/base/fnmatch => ../../../brocade.be/base/fnmatch
	brocade.be/clipboard => ../../../brocade.be/clipboard
	brocade.be/base/fs => ../../../brocade.be/base/fs
	brocade.be/base/mumps => ../../../brocade.be/base/mumps
	brocade.be/base/parallel => ../../../brocade.be/base/parallel
	brocade.be/base/pattern => ../../../brocade.be/base/pattern
	brocade.be/base/python => ../../../brocade.be/base/python
	brocade.be/base/registry => ../../../brocade.be/base/registry
	brocade.be/base/ssh => ../../../brocade.be/base/ssh
	brocade.be/qtechng/cli/cmd => ../../../brocade.be/qtechng/cli/cmd
	brocade.be/qtechng/lib/client => ../../../brocade.be/qtechng/lib/client
	brocade.be/qtechng/lib/error => ../../qtechng/lib/error
	brocade.be/qtechng/lib/file/bfile => ../../qtechng/lib/file/bfile
	brocade.be/qtechng/lib/file/dfile => ../../qtechng/lib/file/dfile
	brocade.be/qtechng/lib/file/ifile => ../../qtechng/lib/file/ifile
	brocade.be/qtechng/lib/file/lfile => ../../qtechng/lib/file/lfile
	brocade.be/qtechng/lib/file/mfile => ../../qtechng/lib/file/mfile
	brocade.be/qtechng/lib/file/objfile => ../../qtechng/lib/file/objfile
	brocade.be/qtechng/lib/file/ofile => ../../qtechng/lib/file/ofile
	brocade.be/qtechng/lib/file/xfile => ../../qtechng/lib/file/xfile
	brocade.be/qtechng/lib/meta => ../../qtechng/lib/meta
	brocade.be/qtechng/lib/object => ../../qtechng/lib/object
	brocade.be/qtechng/lib/project => ../../qtechng/lib/project
	brocade.be/qtechng/lib/server => ../../qtechng/lib/server
	brocade.be/qtechng/lib/source => ../../qtechng/lib/source
	brocade.be/qtechng/lib/util => ../../qtechng/lib/util
	brocade.be/qtechng/lib/vfs => ../../qtechng/lib/vfs
)

require (
	brocade.be/base/fnmatch v0.0.0
	brocade.be/clipboard v0.0.0
	brocade.be/base/fs v0.0.0
	brocade.be/base/mumps v0.0.0
	brocade.be/base/parallel v0.0.0-00010101000000-000000000000
	brocade.be/base/pattern v0.0.0
	brocade.be/base/python v0.0.0
	brocade.be/base/registry v0.0.0
	brocade.be/base/ssh v0.0.0
	brocade.be/qtechng/cli/cmd v0.0.0
	brocade.be/qtechng/lib/client v0.0.0
	brocade.be/qtechng/lib/error v0.0.0
	brocade.be/qtechng/lib/file/bfile v0.0.0
	brocade.be/qtechng/lib/file/dfile v0.0.0
	brocade.be/qtechng/lib/file/ifile v0.0.0
	brocade.be/qtechng/lib/file/lfile v0.0.0
	brocade.be/qtechng/lib/file/mfile v0.0.0
	brocade.be/qtechng/lib/file/objfile v0.0.0
	brocade.be/qtechng/lib/file/ofile v0.0.0
	brocade.be/qtechng/lib/file/xfile v0.0.0
	brocade.be/qtechng/lib/meta v0.0.0
	brocade.be/qtechng/lib/object v0.0.0
	brocade.be/qtechng/lib/project v0.0.0
	brocade.be/qtechng/lib/server v0.0.0
	brocade.be/qtechng/lib/source v0.0.0
	brocade.be/qtechng/lib/util v0.0.0
	brocade.be/qtechng/lib/vfs v0.0.0
	github.com/benhoyt/goawk v1.6.1 // indirect
	github.com/natefinch/atomic v0.0.0-20200526193002-18c0533a5b09 // indirect
	github.com/rwtodd/Go.Sed v0.0.0-20190103233418-906bc69c9394 // indirect
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spyzhov/ajson v0.4.2 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	github.com/zloylos/grsync v0.0.0-20201103111841-a3c772f11d26 // indirect
)
