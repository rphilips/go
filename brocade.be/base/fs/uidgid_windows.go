package fs

import (
	"io/fs"
	"runtime"
)

func uidgid(si fs.FileInfo) (uid int, gid int, ok bool) {
	if runtime.GOOS == "windows" {
		return
	}
	return
}
