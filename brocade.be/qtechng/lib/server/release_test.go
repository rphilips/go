package server

import (
	"testing"
)

func TestRelease(t *testing.T) {
	release, _ := Release{}.New("9.94", false)
	release.FS().RemoveAll("/")
	err := release.Init()
	if err != nil {
		t.Errorf("Creation failed `%s`", err)
	}

	if ok, _ := release.Exists("/source/data"); !ok {
		p, _ := release.FS().RealPath("/source/data")
		t.Errorf("source is not a directory `%s`", p)
	}
	release.FS("").RemoveAll("/")
	if ok, _ := release.Exists("/source/data"); ok {
		p, _ := release.FS().RealPath("/source/data")
		t.Errorf("source is a directory `%s`", p)
	}

}
