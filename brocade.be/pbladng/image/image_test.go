package image

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	pfs "brocade.be/base/fs"
)

func TestImages(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer pfs.Rmpath(tmpdir)
	subdir := path.Join(tmpdir, "prev")
	err = os.Mkdir(subdir, os.ModePerm)
	if err != nil {
		t.Errorf(err.Error())
	}
	for _, p := range []string{"a-A", "b-B", "c-C"} {
		pfs.Store(filepath.Join(tmpdir, p+".jpg"), "", "process")
	}

	for _, p := range []string{"a-A", "b-B", "c-C", "d-D"} {
		pfs.Store(filepath.Join(tmpdir, "prev", p+".jpeg"), "", "process")
	}

	m := ImageMap(tmpdir)

	b, _ := json.MarshalIndent(m, "", "    ")

	if len(m) != 4 {
		t.Errorf(string(b))
		return
	}
	if m["a"] != path.Join("a-A.jpg") {
		t.Errorf(string(b))
		return
	}
	if m["d"] != path.Join("prev", "d-D.jpeg") {
		t.Errorf(string(b))
		return
	}

}

func TestReduce(t *testing.T) {
	imgpath := "t.png"
	tmpdir, _ := ioutil.TempDir("", "")
	work := filepath.Join(tmpdir, imgpath)
	pfs.CopyFile(imgpath, work, "process", false)
	err := ReduceSize(work, 500)
	if err != nil {
		t.Errorf(err.Error())
	}
}
