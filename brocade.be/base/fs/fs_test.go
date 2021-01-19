package fs

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"brocade.be/base/registry"
)

func TestPathURI(t *testing.T) {

	path := "/laptop/My Documents/FileSchemeURIs.doc"
	uri, err := PathURI(path)

	if err != nil {
		t.Errorf("fs.PathURI fails")
	}
	if err == nil {
		if uri != "file:///laptop/My%20Documents/FileSchemeURIs.doc" {
			t.Errorf("Should be ok, found %s", uri)
		}
	}
}

func TestAbsPath(t *testing.T) {

	path, err := AbsPath("$BROCADE_REGISTRY")
	_, err = ioutil.ReadFile(path)

	if err != nil {
		t.Errorf("fs.AbsPath fails")
	}

	path, err = AbsPath("$BROCADE_REGISTRY/shouldnotexist")
	_, err = ioutil.ReadFile(path)
	if err == nil {
		t.Errorf("fs.AbsPath fails")
	}
}

func TestProperties(t *testing.T) {
	perm, err := Properties("qtechfile")
	if err != nil {
		t.Errorf("Should be ok")
	}
	if perm.PERM != 0664 {
		t.Errorf("mode should be '0664'")
	}

	_, err = Properties("abcdefg")
	if err != ErrNotPathMode {
		t.Errorf("Should give error")
	}
}

func TestStore(t *testing.T) {
	write := "Hello World"
	filename := "test.txt"
	err := Store(filename, write, "process")
	if err != nil {
		t.Errorf("Should succeed: %v", err)
	}
	err = Store(filename, write, "qtech")
	if err != nil {
		t.Errorf("Should succeed")
	}
	data, err := Fetch(filename)
	if string(data) != write {
		t.Errorf("Should succeed 2")
	}
}

func TestMkdir(t *testing.T) {
	dirname := filepath.Join(registry.Registry["scratch-dir"], "Hello", "World")
	_ = Rmpath(dirname)
	err := Mkdir(dirname, "process")
	if err != nil {
		t.Errorf("Cannot make dir")
	}
	write := "Hello World"
	filename := filepath.Join(dirname, "test.txt")
	err = Store(filename, write, "process")
	data, err := Fetch(filename)
	if string(data) != write {
		t.Errorf("Should succeed 2")
	}
	err = Rmpath(dirname)
	if err != nil {
		t.Errorf("Cannot remove dir")
	}
}

func TestCopyFile(t *testing.T) {

	scratchdir := registry.Registry["scratch-dir"]
	fname := filepath.Join(scratchdir, "Hello")
	Store(fname, "World", "process")
	CopyFile(fname, fname+"2", "=", true)
	CopyFile(fname, fname+"3", "", true)
}
