package project

import (
	"encoding/json"
	"fmt"
	"testing"

	qmeta "brocade.be/qtechng/lib/meta"
	qserver "brocade.be/qtechng/lib/server"
)

func TestProject01(t *testing.T) {
	release := makeRelease()
	r := release.String()

	proj := "/a/b/c/d"
	project, err := Project{}.New(r, proj, false)

	project.Init(qmeta.Meta{})
	if err != nil {
		t.Errorf(fmt.Sprintf("error is `%s`", err))
		return
	}

	config := `{"py3": true,"notunique": ["A.m", "B.m"]}`
	_, err = project.Store("brocade.json", []byte(config))
	if err != nil {
		t.Errorf(fmt.Sprintf("error is `%s`", err))
		return
	}

	cfg, err := project.LoadConfig()

	if err != nil {
		t.Errorf(fmt.Sprintf("error on load `%s`", err))
		return
	}
	if !cfg.Py3 {
		t.Errorf(fmt.Sprintf("on load `%v`", cfg))
		return
	}
	if len(cfg.NotUnique) != 2 {
		t.Errorf(fmt.Sprintf("on load `%v`", cfg))
		return
	}

}

func TestCore01(t *testing.T) {
	release := makeRelease()
	r := release.String()

	for _, proj := range []string{"/a/b", "/a/b/c/d"} {
		project, err := Project{}.New(r, proj, false)
		if err != nil {
			t.Errorf(fmt.Sprintf("error is `%s`", err))
			return
		}
		project.Init(qmeta.Meta{})
	}
	project, _ := Project{}.New(r, "/a/b", true)
	if project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should not be core: %s", project.String()))
		return
	}

	project, _ = Project{}.New(r, "/a/b/c/d", true)
	if project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should not be core: %s", project.String()))
		return
	}

	project, _ = Project{}.New(r, "/a/b/c", true)
	if project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should not be core: %s", project.String()))
		return
	}

}

func TestCore02(t *testing.T) {
	release := makeRelease()
	r := release.String()

	for _, proj := range []string{"/a/b", "/a/b/c/d"} {
		project, err := Project{}.New(r, proj, false)
		if err != nil {
			t.Errorf(fmt.Sprintf("error is `%s`", err))
			return
		}
		project.Init(qmeta.Meta{})
		if proj == "/a/b/c/d" {
			continue
		}
		config := `{"core": true,"priority":3333}`
		project.Store("brocade.json", []byte(config))
		cfg := Config{}
		json.Unmarshal([]byte(config), &cfg)
		project.UpdateConfig(cfg)

	}

	project, _ := Project{}.New(r, "/a/b", true)
	if !project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should be core: %s", project.String()))
		return
	}
	sort := project.Orden()
	if sort != "1996667" {
		t.Errorf(fmt.Sprintf("sort: %s", sort))
		return
	}

	project, _ = Project{}.New(r, "/a/b/c/d", true)
	if !project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should be core: %s", project.String()))
		return
	}

	project, _ = Project{}.New(r, "/a/b/c", true)
	if project.IsCore() {
		t.Errorf(fmt.Sprintf("Project should not be core: %s", project.String()))
		return
	}
}

func TestSeq01(t *testing.T) {
	release := makeRelease()
	r := release.String()

	for _, proj := range []string{"/a/b", "/a/b/c/d"} {
		project, err := Project{}.New(r, proj, false)
		if err != nil {
			t.Errorf(fmt.Sprintf("error is `%s`", err))
			return
		}
		project.Init(qmeta.Meta{})
		if proj == "/a/b/c/d" {
			continue
		}
		config := `{"core": true}`
		project.Store("brocade.json", []byte(config))
		cfg := Config{}
		json.Unmarshal([]byte(config), &cfg)
		project.UpdateConfig(cfg)
	}
	seq1, _ := Sequence(r, "/a/b", true)
	seq11, _ := Sequence(r, "/a/b/brocade.json", true)
	seq12, _ := Sequence(r, "/a/b/brocade1.json", true)

	if len(seq1) != 1 {
		t.Errorf(fmt.Sprintf("\n%v", seq1))
		return
	}
	if len(seq11) != 1 {
		t.Errorf(fmt.Sprintf("\n%v", seq11))
		return
	}
	if len(seq12) != 1 {
		t.Errorf(fmt.Sprintf("\n%v", seq11))
		return
	}

	seq3, _ := Sequence(r, "/a/b/c/d", true)

	if len(seq3) != 2 {
		t.Errorf(fmt.Sprintf("\n%v", seq3))
		return
	}

	seq2, _ := Sequence(r, "/a/b/c", true)

	if len(seq2) != 1 {
		t.Errorf(fmt.Sprintf("\n%v", seq2))
		return
	}

	if seq1[0] != seq2[0] {
		t.Errorf(fmt.Sprintf("\n%v", seq2))
		return
	}

}

func makeRelease() (release *qserver.Release) {
	release, _ = qserver.Release{}.New("9.94", false)
	release.FS("/").RemoveAll("/")
	release.Init()
	return release
}
