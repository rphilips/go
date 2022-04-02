package python

import (
	"path"
	"testing"

	qfs "brocade.be/base/fs"
)

func TestCompile1(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptpy := path.Join(dir, "script.py")

	if Compile(scriptpy, true, false, nil) == nil {
		t.Errorf("py3 compilation should fail!")
	}

	if Compile(scriptpy, false, false, nil) == nil {
		t.Errorf("py2 compilation should fail!")
	}
}

func TestCompile2(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptpy := path.Join(dir, "script.py")
	qfs.Store(scriptpy, "pass", "process")

	if Compile(scriptpy, true, false, nil) != nil {
		t.Errorf("py3 compilation should not fail!")
	}

	if Compile(scriptpy, false, false, nil) != nil {
		t.Errorf("py2 compilation should not fail!")
	}
}

func TestCompile3(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptpy := path.Join(dir, "script.py")
	qfs.Store(scriptpy, "print 1", "process")

	if Compile(scriptpy, true, false, nil) == nil {
		t.Errorf("py3 compilation should fail!")
	}

	if Compile(scriptpy, false, false, nil) != nil {
		t.Errorf("py2 compilation should not fail!")
	}
}

func TestCompile4(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptpy := path.Join(dir, "script.py")
	qfs.Store(scriptpy, "print(1)", "process")

	if Compile(scriptpy, true, false, nil) != nil {
		t.Errorf("py3 compilation should not fail!")
	}

	if Compile(scriptpy, false, false, nil) != nil {
		t.Errorf("py2 compilation should not fail!")
	}
}
