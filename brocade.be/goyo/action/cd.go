package action

import (
	"fmt"
	"os"
	"path/filepath"

	qfs "brocade.be/base/fs"
	qutil "brocade.be/goyo/lib/util"
)

func Cd(text string) []string {

	home, _ := os.UserHomeDir()
	dir := ""
	switch text {
	case "":
		desktop := filepath.Join(home, "Desktop")
		if qfs.IsDir(desktop) {
			dir = desktop
		} else {
			dir = home
		}
	default:
		d, e := qfs.AbsPath(text)
		if e != nil {
			dir = text
		} else {
			dir = d
		}
	}

	err := os.Chdir(dir)
	if err == nil {
		cwd, _ := os.Getwd()
		fmt.Println(cwd)
		return []string{"cd " + cwd}
	} else {
		qutil.Error(err)
		return nil
	}
}
