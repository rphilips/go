package fs

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	path, _ := AbsPath("$BROCADE_REGISTRY")
	_, err := os.ReadFile(path)

	if err != nil {
		t.Errorf("fs.AbsPath fails")
	}

	path, _ = AbsPath("$BROCADE_REGISTRY/shouldnotexist")
	_, err = os.ReadFile(path)
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
	data, _ := Fetch(filename)
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
	Store(filename, write, "process")
	data, _ := Fetch(filename)
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

func TestFind(t *testing.T) {
	files := []string{
		"/a/f.pdf",
		"/a/f.txt",
		"/a/g.txt",
		"/a/b/ff.txt",
		"/a/b/ff.pdf",
		"/a/b/c/fff.pdf",
	}
	tmpdir, _ := TempDir("", "tst")
	for _, fname := range files {
		parts := strings.SplitN(fname, "/", -1)
		parts[0] = tmpdir
		p := path.Join(parts...)
		dirname := filepath.Dir(p)
		os.MkdirAll(dirname, os.ModePerm)
		Store(p, "", "")
	}

	matches, err := Find(tmpdir, []string{"f*"}, true, true, false)

	if err != nil {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
		return
	}
	if len(matches) != 5 {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
	}
	matches, err = Find(tmpdir, []string{"g*", "fff*.pdf"}, true, true, false)
	if err != nil {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
		return
	}
	if len(matches) != 2 {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
	}

	matches, err = Find(tmpdir, []string{"[bf]*"}, true, true, true)
	if err != nil {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
		return
	}
	if len(matches) != 6 {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
	}

	matches, err = Find(tmpdir, []string{"[bf]*"}, true, false, true)
	if err != nil {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
		return
	}
	if len(matches) != 1 {
		t.Errorf("\nFail:\n%s\n%s\n", tmpdir, err)
		fmt.Println(matches)
	}

}

func TestCalcPerm(t *testing.T) {
	type tst struct {
		readable string
		number   os.FileMode
	}
	tests := []tst{
		{readable: "rwxrwxrwx", number: 0777},
		{readable: "rwsrwxrwx", number: 04777},
		{readable: "rw-rw-r--", number: 436},
	}

	for _, test := range tests {
		expected := test.number
		calc := CalcPerm(test.readable)
		if expected == calc {
			continue
		}
		t.Errorf("\nFail:\n%s\n%s\n%s\n%d\n", test.readable, expected, calc, int64(calc))
	}
	fname := "/home/rphilips/tmp/a"
	props, _ := Properties("webfile")
	perm := props.PERM
	os.Chmod(fname, perm)
}

func TestChangedAfter(t *testing.T) {
	rootdir := "/home/rphilips/tmp"
	skip := []string{"a", "b/c", "b/c/d/"}
	reffile := "/home/rphilips/Desktop/time.ref"
	after, _ := GetMTime(reffile)
	paths, err := ChangedAfter(rootdir, after, skip)
	t.Errorf("error: %v\n", err)
	t.Errorf("paths: %v\n", paths)
	t.Errorf("after: %v\n", after)
}
