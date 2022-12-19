package document

import (
	"fmt"
	"os"
	"path/filepath"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
)

func Archive(dir string) (err error) {
	if dir == "" {
		dir = pfs.FName("workspace")
	}

	year, week, err := DocRef(dir)
	if err != nil {
		err = fmt.Errorf("cannot deduce previous year and week: %s", err)
		return
	}
	archivedir := pfs.FName(fmt.Sprintf("archive/%d/%02d", year, week))

	files, err := os.ReadDir(dir)
	if err != nil {
		err = fmt.Errorf("cannot list files in `%s`: %s", dir, err)
	}

	for _, file := range files {
		sfile := filepath.Join(dir, file.Name())
		if !file.IsDir() {
			tfile := filepath.Join(archivedir, file.Name())
			err = bfs.CopyFile(sfile, tfile, "", false)
			if err != nil {
				err = fmt.Errorf("cannot copy `%s` to `%s`", sfile, tfile)
				return
			}
		}
		if file.Name() != "week.md" {
			bfs.Rmpath(sfile)
		}
	}
	return
}
