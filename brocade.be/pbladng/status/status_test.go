package status

import (
	"io/ioutil"
	"testing"

	pfs "brocade.be/base/fs"
)

func TestInit(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer pfs.Rmpath(tmpdir)
	status := Status{}

	(&status).Init(tmpdir, 2022, 13)

	if status.Year != 2022 {
		t.Errorf("Wrong year")
	}
	if status.Week != 13 {
		t.Errorf("Wrong month")
	}

	(&status).Init(tmpdir, 2022, 0)
	if status.Year != 2022 {
		t.Errorf("Wrong year")
	}
	if status.Week != 13 {
		t.Errorf("Wrong month")
	}
	t.Errorf(tmpdir)
}
