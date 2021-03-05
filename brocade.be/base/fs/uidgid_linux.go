package fs

import (
	"io/fs"
	"runtime"
	"syscall"
)

func uidgid(si fs.FileInfo) (uid int, gid int, ok bool) {
	if runtime.GOOS == "windows" {
		return
	}
	if stat, ok := si.Sys().(*syscall.Stat_t); ok {
		uid = int(stat.Uid)
		gid = int(stat.Gid)
		return uid, gid, ok
	}
	return
}
