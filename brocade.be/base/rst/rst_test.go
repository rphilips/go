package rst

import (
	"path"
	"testing"

	qfs "brocade.be/base/fs"
)

func TestCheck1(t *testing.T) {
	dir, _ := qfs.TempDir("", "")
	scriptrst := path.Join(dir, "script.rst")
	qfs.Store(scriptrst, "*Hello", "qtech")

	if Check(scriptrst, "") == nil {
		t.Errorf("rst is not ok!: %s", scriptrst)
	}

}
