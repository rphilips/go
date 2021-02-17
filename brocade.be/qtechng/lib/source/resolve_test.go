package source

import (
	"os/exec"
	"path/filepath"
	"sort"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"

	qfs "brocade.be/base/fs"
)

func TestResolv01(t *testing.T) {
	tmpdir, err := setupResolve()
	if err != nil {
		t.Errorf("\nAt %s:\n\n%s", tmpdir, err.Error())
		return
	}
	tests, _ := listTxt(tmpdir)
	sort.Strings(tests)

	for _, test := range tests {
		err := rsRS(test)
		if err != nil {
			t.Errorf("\nAt %s: %s\n\n%s", tmpdir, test, err.Error())
			return
		}
		testok := test + ".ok"

		blobok, err := qfs.Fetch(testok)
		if err != nil {
			t.Errorf("\nAt %s: %s\n: %s", tmpdir, testok, err.Error())
			continue
		}
		testrs := test + ".rs"
		blobrs, err := qfs.Fetch(testrs)
		if err != nil {
			t.Errorf("\nAt %s: %s\n: %s", tmpdir, testrs, err.Error())
			continue
		}

		if string(blobrs) != string(blobok) {
			t.Errorf("\nAt %s: %s\n\n    %s\n    %s\n%s", tmpdir, filepath.Base(test), filepath.Base(testok), filepath.Base(testrs), diff(blobok, blobrs))
			continue
		}

	}
}

func diff(blobok, blobrs []byte) string {

	dmp := diffmatchpatch.New()
	d := dmp.DiffMain(string(blobok), string(blobrs), false)

	s := dmp.DiffPrettyText(d)

	// for _, x := range []string{"\x1b[32m", "\x1b[0m", "\x1b[31m"} {
	// 	s = strings.ReplaceAll(s, x, "\n===\n")
	// }

	return s
}

func setupResolve() (tmpdir string, err error) {
	tmpdir, err = qfs.TempDir("", "qt")
	if err != nil {
		return
	}
	err = coRS("/qtechng/test", tmpdir)
	return
}

func coRS(project string, tmpdir string) (err error) {
	cmd := exec.Command("qtechng", "project", "co", "/qtechng/test", "--quiet")
	cmd.Dir = tmpdir
	return cmd.Run()
}

func listTxt(tmpdir string) (paths []string, err error) {
	paths, err = qfs.Find(tmpdir, []string{"*.txt"}, false, true, false)
	return
}

func rsRS(path string) (err error) {
	qpath := "/qtechng/test/" + filepath.Base(path)
	rspath := path + ".rs"
	qfs.Rmpath(rspath)
	cmd := exec.Command("qtechng", "source", "resolve", qpath, "--stdout="+rspath)
	return cmd.Run()
}
