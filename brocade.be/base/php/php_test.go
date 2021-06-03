package php

import (
	"path"
	"testing"

	qfs "brocade.be/base/fs"
)

func TestCompile1(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptphp := path.Join(dir, "script.php")

	if Compile(scriptphp) == nil {
		t.Errorf("php compilation should fail!")
	}

}
