package source

import (
	"fmt"
	"testing"

	qmeta "brocade.be/qtechng/lib/meta"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
)

func TestNature01(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/install.py"
	proj := "/a/b/c"
	release, _ := makenRelease(r, proj)
	r = release.String()

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	natures := source.Natures()

	if len(natures) != 2 {
		t.Errorf("Should have 2 natures !: %v", natures)
		return
	}
	if !natures["text"] {
		t.Errorf("Should have nature text !")
		return
	}
	if !natures["install"] {
		t.Errorf("Should have nature install !: %v", natures)
		return
	}

}

func TestNature02(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/macro.d"
	proj := "/a/b/c"
	release, _ := makenRelease(r, proj)
	r = release.String()

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	natures := source.Natures()

	if len(natures) != 4 {
		t.Errorf("Should have 4 natures !: %v", natures)
		return
	}
	if !natures["text"] {
		t.Errorf("Should have nature text !: %v", natures)
		return
	}
	if !natures["auto"] {
		t.Errorf("Should have nature auto !: %v", natures)
		return
	}
	if !natures["dfile"] {
		t.Errorf("Should have nature `d` !: %v", natures)
		return
	}
	if !natures["objectfile"] {
		t.Errorf("Should have nature `o` !: %v", natures)
		return
	}

}

func TestNature021(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/macro.l"
	proj := "/a/b/c"
	release, _ := makenRelease(r, proj)
	r = release.String()

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	natures := source.Natures()

	if len(natures) != 5 {
		t.Errorf("Should have 5 natures !: %v", natures)
		return
	}
	if !natures["text"] {
		t.Errorf("Should have nature text !: %v", natures)
		return
	}
	if !natures["auto"] {
		t.Errorf("Should have nature auto !: %v", natures)
		return
	}
	if !natures["lfile"] {
		t.Errorf("Should have nature `lfile` !: %v", natures)
		return
	}
	if !natures["objectfile"] {
		t.Errorf("Should have nature `objectfile` !: %v", natures)
		return
	}

	if !natures["mumps"] {
		t.Errorf("Should have nature `mumps` !: %v", natures)
		return
	}

}

func TestNature03(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/macro.d"
	proj := "/a/b/c"
	release, _ := makenRelease(r, proj)
	r = release.String()

	cfg := `{"notbrocade":["macro.d"]}`

	scfg, _ := Source{}.New(r, "/a/b/c/brocade.json", false)
	scfg.Store(qmeta.Meta{}, cfg, false)
	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	natures := source.Natures()

	if len(natures) != 1 {
		t.Errorf("Should have 1 natures !: %v", natures)
		return
	}
	if !natures["text"] {
		t.Errorf("Should have nature text !: %v", natures)
		return
	}

}

func TestNature04(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/d/brocade.json"
	proj := "/a/b/c"
	release, _ := makenRelease(r, proj)
	r = release.String()

	cfg := `{"notconfig":["d/brocade.json"], "binary":["d/brocade.json"]}`

	scfg, _ := Source{}.New(r, "/a/b/c/brocade.json", false)
	scfg.Store(qmeta.Meta{}, cfg, false)
	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	natures := source.Natures()

	if len(natures) != 1 {
		t.Errorf("Should have 1 natures !: %v", natures)
		return
	}
	if !natures["binary"] {
		t.Errorf("Should have nature binary !: %v", natures)
		return
	}

}

func makenRelease(r string, proj string) (release *qserver.Release, project *qproject.Project) {
	release, _ = qserver.Release{}.New(r, false)
	release.FS("/").RemoveAll("/")
	err := release.Init()
	if err != nil {
		fmt.Println("error server:", err)
	}

	if proj != "" {
		project, _ = qproject.Project{}.New(r, proj, false)
		project.Init(qmeta.Meta{})
	}
	return
}
