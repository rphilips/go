package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"

	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

func TestSource01(t *testing.T) {
	r := "9.99"
	proj := ""
	release, _ := makeRelease(r, proj)
	r = release.String()

	p := "/a/b/c/hallo.txt"

	_, err := Source{}.New(r, p, false)

	if err == nil {
		t.Errorf("No Project\n")
		return
	}
	ref := err.(*qerror.QError).Ref
	if ref[0] != "source.new.noproject" {
		t.Errorf(fmt.Sprintf("No Project\n%s", err.Error()))
		return
	}

	proj = "/a/b/c"
	_, _ = makeRelease(r, proj)

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf("Source should be created\n")
		return
	}
	if source.String() != p {
		t.Errorf(fmt.Sprintf("Source should be:\n" + p))
		return
	}
}

func TestSource02(t *testing.T) {
	r := "9.99"
	p := "/a/b/c/hallo.txt"
	proj := "/a/b/c"
	release, _ := makeRelease(r, proj)
	r = release.String()

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	_, _, _, err = source.Store(qmeta.Meta{}, "Hello World", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if source.String() != p {
		t.Errorf("\nstring: %s\npath: %s\n", source.String(), source.Path())
		return
	}
	content, err := source.Fetch()

	if err != nil {
		t.Errorf(err.Error())
		return
	}

	if string(content) != "Hello World" {
		t.Errorf("Found: %s", string(content))
		return
	}

}

func TestSource03(t *testing.T) {
	r := "9.98"
	p := "/a/b/c/hallo.txt"
	proj := "/a/b/c"
	release, _ := makeRelease(r, proj)
	r = release.String()

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	mt := qmeta.Meta{Cu: "nu"}
	_, _, _, err = source.Store(mt, "Hello World", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	_, err = source.Fetch()

	if err != nil {
		t.Errorf(err.Error())
		return
	}
	met1, err := qmeta.Meta{}.New(r, p)
	met0 := *met1
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	if met1.Cu != "nu" || met1.Mt != met1.Ct {
		t.Errorf("met1 %v", met1)
		t.Errorf("mt %v", mt)
		return
	}
	fmt.Printf("\nmet1 %p %v\n", met1, met1)

	time.Sleep(1000 * time.Millisecond)

	met3, _, _, err := source.Store(mt, "Hello World2", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	met2, _ := qmeta.Meta{}.New(r, p)
	if met2.Cu != "nu" || met2.Mt == met2.Ct || met2.Mt == met0.Mt {
		t.Errorf("met2 %p %v", met2, met2)
		t.Errorf("met1 %p %v", met1, met1)
		t.Errorf("met3 %p %v", met3, met3)
		t.Errorf("mt %v", mt)
		return
	}
}

func TestSource04(t *testing.T) {
	r := "9.99"
	p1 := "/a/b/c/hallo.ext1"
	proj1 := "/a/b/c"
	release, _ := makeRelease(r, proj1)
	r = release.String()

	qregistry.Registry["qtechng-unique-ext"] = ".ext2"

	proj2 := "/a2"
	if proj2 != "" {
		project2, _ := qproject.Project{}.New(r, proj2, false)
		project2.Init(qmeta.Meta{})
	}

	source1, err := Source{}.New(r, p1, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	p2 := "/a2/hallo.ext1"
	source2, err := Source{}.New(r, p2, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	_, _, _, err = source1.Store(qmeta.Meta{}, "Hello World1", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	_, _, _, err = source2.Store(qmeta.Meta{}, "Hello World2", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	qregistry.Registry["qtechng-unique-ext"] = ".ext1"

	_, _, _, err = source2.Store(qmeta.Meta{}, "Hello World2", false)
	if err == nil {
		t.Errorf("Should check on uniqueness")
		return
	}

}

func TestSource05(t *testing.T) {
	r := "9.99"
	p1 := "/a/b/c/hallo.ext1"
	proj1 := "/a/b/c"
	release, _ := makeRelease(r, proj1)
	r = release.String()

	qregistry.Registry["qtechng-unique-ext"] = ".ext1"

	source1, err := Source{}.New(r, p1, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	_, _, _, err = source1.Store(qmeta.Meta{}, "Hello World3", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	_, _, _, err = source1.Store(qmeta.Meta{}, "Hello World2", false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	release.FS().RemoveAll("/")
}

func TestSource06(t *testing.T) {
	// brocade.json
	r := "9.99"
	bjson := "/a/b/brocade.json"
	proj1 := "/a/b"
	release, _ := makeRelease(r, proj1)
	r = release.String()
	blob := []byte(`{"core": 1}`)

	source, _ := Source{}.New(r, bjson, false)
	_, _, _, err := source.Store(qmeta.Meta{}, blob, false)

	if err == nil {
		t.Errorf("Should show an error: %v", source.project.IsConfig(source.String()))
		return
	}

	blob = []byte(`{"core": true}`)

	source, _ = Source{}.New(r, bjson, false)
	_, _, _, err = source.Store(qmeta.Meta{}, blob, false)

	if err != nil {
		t.Errorf(err.Error())
		return
	}
}

// /usr/local/go/bin/go test -timeout 500s brocade.be/qtechng/lib/source -run "^(TestSourceList01)$"
// ok  	brocade.be/qtechng/lib/source	127.515s
// ok  	brocade.be/qtechng/lib/source	191.133s

func TestSourceList01(t *testing.T) {
	// brocade.json
	r := "9.99"
	release, _ := makeRelease(r, "")
	r = release.String()

	// creates projects
	paths := make([]string, 0)
	projects := make([]string, 0)
	prun := 100

	for i := 0; i < prun; i++ {
		p := "/a" + strconv.Itoa(i) + "/b" + strconv.Itoa(i)
		projects = append(projects, p)
		for j := 0; j < 100; j++ {
			s := p + "/f" + strconv.Itoa(j) + ".txt"
			paths = append(paths, s)
		}
		paths = append(paths, p+"/brocade.json")
	}
	_, errs := qproject.InitList(r, projects, func(p string) qmeta.Meta { return qmeta.Meta{} })

	if errs != nil {
		t.Errorf(errs.Error())
		return
	}

	// ok
	data := strings.Repeat("abcd", 10000)
	fmeta := func(p string) qmeta.Meta { return qmeta.Meta{} }
	fdata := func(p string) ([]byte, error) {
		if strings.HasSuffix(p, "/brocade.json") {
			d := `{"core": false}`
			return []byte(d), nil
		}
		digest := qutil.Digest([]byte(p))
		first := digest[:1]
		mode := "m4"
		switch first {
		case "1", "2", "3":
			mode = "m4"
		case "4", "5":
			mode = "i4"
		case "6", "7", "8", "9", "0":
			mode = "r4"
		default:
			mode = "l4"
		}
		obj := mode + "_z" + digest
		return []byte(data + " " + obj + " " + p), nil
	}

	results, es := StoreList("install", r, paths, false, fmeta, fdata)
	if es != nil {
		t.Errorf(es.Error())
		return
	}

	if len(results) != len(paths) {
		t.Errorf("Not enough paths: \n\n%d\n\n%v\n\n%d\n\n%v\n", len(paths), paths, len(results), results)
		return
	}

	for p := range results {
		if results[p] == nil || results[p].Digest == "" {
			t.Errorf("Should be changed: " + p)
			return
		}
	}

	es = WasteList(r, paths)

	if es != nil {
		t.Errorf(es.Error())
		return
	}

	fs := release.FS("/")
	for _, dir := range []string{"/source/data", "/meta", "/unique", "/tmp", "/object/m4", "/object/l4", "/object/i4"} {
		if ok, _ := fs.Exists(dir); !ok {
			t.Errorf("%s should exist", dir)
			return
		}
		files := fs.Dir(dir, false, false)
		if len(files) > 0 {
			t.Errorf("%s should be empty", dir)
			return
		}
	}

}

func TestSourceObject01(t *testing.T) {
	// brocade.json
	r := "9.99"
	release, _ := makeRelease(r, "")
	r = release.String()

	// creates projectss
	projects := []string{"/a/b"}
	_, errs := qproject.InitList(r, projects, func(p string) qmeta.Meta { return qmeta.Meta{} })

	f1 := "/a/b/f1.txt"
	data1 := "m4_A 1 m4_B m4_C"
	f2 := "/a/b/f2.txt"
	data2 := "m4_A 2 m4_B"
	f3 := "/a/b/f3.txt"
	data3 := "3!"

	fmeta := func(p string) qmeta.Meta { return qmeta.Meta{} }
	fdata := func(p string) ([]byte, error) {
		switch p {
		case f1:
			return []byte(data1), nil
		case f2:
			return []byte(data2), nil
		default:
			return []byte(data3), nil
		}
	}
	_, es := StoreList("install", r, []string{f1, f2, f3}, false, fmeta, fdata)
	if es != nil {
		t.Errorf(errs.Error())
		return
	}
	files := qobject.GetDependencies(release, "m4_A")["m4_A"]
	if len(files) != 2 {
		t.Errorf("Number of files should be 2: %v", files)
		return
	}
	if files[0] != f1 && files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}
	if files[1] != f1 && files[1] != f2 {
		t.Errorf("Bad file2: %s", files[0])
		return
	}

	data1 = "1 m4_B"
	_, es = StoreList("install", r, []string{f1}, false, fmeta, fdata)
	if es != nil {
		t.Errorf(errs.Error())
		return
	}

	files = qobject.GetDependencies(release, "m4_B")["m4_B"]
	if len(files) != 2 {
		t.Errorf("Number of files should be 2: %v", files)
		return
	}
	if files[0] != f1 && files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}
	if files[1] != f1 && files[1] != f2 {
		t.Errorf("Bad file2: %s", files[0])
		return
	}

	files = qobject.GetDependencies(release, "m4_A")["m4_A"]
	if len(files) != 1 {
		t.Errorf("Number of files should be 1: %v", files)
		return
	}
	if files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}

	files = qobject.GetDependencies(release, "m4_C")["m4_C"]
	if len(files) != 0 {
		t.Errorf("Number of files should be 0: %v", files)
		return
	}

	es = WasteList(r, []string{f1, f2, f3})

	if es != nil {
		t.Errorf(es.Error())
		return
	}

	es = WasteList(r, []string{"/a/b/brocade.json"})
	if es != nil {
		t.Errorf(es.Error())
		return
	}

	fs := release.FS("/")
	for _, dir := range []string{"/source/data", "/meta", "/unique", "/tmp", "/object/m4", "/object/l4", "/object/i4"} {
		if ok, _ := fs.Exists(dir); !ok {
			t.Errorf("%s should exist", dir)
			return
		}
		files := fs.Dir(dir, false, false)
		if len(files) > 0 {
			t.Errorf("%s should be empty", dir)
			return
		}
	}

}

func TestSourceObject02(t *testing.T) {
	// brocade.json
	r := "9.99"
	release, _ := makeRelease(r, "")
	r = release.String()

	// creates projectss
	projects := []string{"/a/b"}
	_, errs := qproject.InitList(r, projects, func(p string) qmeta.Meta { return qmeta.Meta{} })

	f1 := "/a/b/f1.txt"
	data1 := "m4_A 1 m4_B m4_C"
	f2 := "/a/b/f2.txt"
	data2 := "m4_A 2 m4_B"
	f3 := "/a/b/f3.txt"
	data3 := "3!"

	fmeta := func(p string) qmeta.Meta { return qmeta.Meta{} }
	fdata := func(p string) ([]byte, error) {
		switch p {
		case f1:
			return []byte(data1), nil
		case f2:
			return []byte(data2), nil
		default:
			return []byte(data3), nil
		}
	}
	_, es := StoreList("install", r, []string{f1, f2, f3}, false, fmeta, fdata)
	if es != nil {
		t.Errorf(errs.Error())
		return
	}
	files := qobject.GetDependencies(release, "m4_A")["m4_A"]
	if len(files) != 2 {
		t.Errorf("Number of files should be 2: %v", files)
		return
	}
	if files[0] != f1 && files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}
	if files[1] != f1 && files[1] != f2 {
		t.Errorf("Bad file2: %s", files[0])
		return
	}

	data1 = "1 m4_B"
	_, es = StoreList("install", r, []string{f1}, false, fmeta, fdata)
	if es != nil {
		t.Errorf(errs.Error())
		return
	}

	files = qobject.GetDependencies(release, "m4_B")["m4_B"]
	if len(files) != 2 {
		t.Errorf("Number of files should be 2: %v", files)
		return
	}
	if files[0] != f1 && files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}
	if files[1] != f1 && files[1] != f2 {
		t.Errorf("Bad file2: %s", files[0])
		return
	}

	files = qobject.GetDependencies(release, "m4_A")["m4_A"]
	if len(files) != 1 {
		t.Errorf("Number of files should be 1: %v", files)
		return
	}
	if files[0] != f2 {
		t.Errorf("Bad file1: %s", files[0])
		return
	}

	files = qobject.GetDependencies(release, "m4_C")["m4_C"]
	if len(files) != 0 {
		t.Errorf("Number of files should be 0: %v", files)
		return
	}

	es = WasteList(r, []string{f1, f2, f3})

	if es != nil {
		t.Errorf(es.Error())
		return
	}

	es = WasteList(r, []string{"/a/b/brocade.json"})
	if es != nil {
		t.Errorf(es.Error())
		return
	}

	fs := release.FS("/")
	for _, dir := range []string{"/source/data", "/meta", "/unique", "/tmp", "/object/m4", "/object/l4", "/object/i4"} {
		if ok, _ := fs.Exists(dir); !ok {
			t.Errorf("%s should exist", dir)
			return
		}
		files := fs.Dir(dir, false, false)
		if len(files) > 0 {
			t.Errorf("%s should be empty", dir)
			return
		}
	}

}

func TestSourceTree01(t *testing.T) {
	// brocade.json
	r := "9.98"
	release, _ := makeRelease(r, "")
	r = release.String()

	fmeta := func(p string) qmeta.Meta { return qmeta.Meta{} }
	basedir := "/media/rphilips/SAN2/brocade"
	_, errs := StoreTree("tree", r, basedir, fmeta)

	if errs != nil {
		t.Errorf(errs.Error())
		return
	}
}

func TestSourceObj01(t *testing.T) {
	// Store a d-file
	r := "9.99"
	proj := "/a/b/c"
	release, _ := makeRelease(r, proj)
	r = release.String()

	testfile := filepath.Join(qregistry.Registry["qtechng-test-dir"], "cat.d")
	data, _ := os.ReadFile(testfile)

	p := proj + "/acat.d"
	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	_, _, _, err = source.Store(qmeta.Meta{}, data, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

}

func TestSourceObj02(t *testing.T) {
	// brocade.json
	r := "9.99"
	proj := "/a/b/c"
	release, _ := makeRelease(r, proj)
	r = release.String()

	data := dfile1()

	p := proj + "/acat.d"
	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	_, _, _, err = source.Store(qmeta.Meta{}, data, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	dep1 := qobject.GetDependencies(release, "m4_getCatGenStatus")["m4_getCatGenStatus"]
	if len(dep1) != 1 || dep1[0] != "/a/b/c/acat.d" {
		t.Errorf("Wrong dependencies: %s", dep1)
	}

	dep2 := qobject.GetDependencies(release, "m4_CO")["m4_CO"]
	if len(dep2) != 1 || dep2[0] != "m4_setCatGenStatus" {
		t.Errorf("Wrong dependencies: %s", dep2)
	}

	data = dfile2()

	_, _, _, err = source.Store(qmeta.Meta{}, data, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	dep1 = qobject.GetDependencies(release, "m4_setCatGenStatus")["m4_setCatGenStatus"]
	if len(dep1) != 1 || dep1[0] != "/a/b/c/acat.d" {
		t.Errorf("Wrong dependencies: %s", dep1)
	}

	dep2 = qobject.GetDependencies(release, "m4_CO")["m4_CO"]
	if len(dep2) != 0 {
		t.Errorf("Wrong dependencies: %s", dep2)
	}

	p = proj + "/my.txt"
	tsource, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	_, _, _, err = tsource.Store(qmeta.Meta{}, dfile3(), false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	dep1 = qobject.GetDependencies(release, "m4_getCatGenStatus")["m4_getCatGenStatus"]
	if len(dep1) != 2 {
		t.Errorf("Wrong dependencies: %s", dep1)
	}
	sort.Strings(dep1)
	if dep1[0] != "/a/b/c/acat.d" || dep1[1] != "/a/b/c/my.txt" {
		t.Errorf("Wrong dependencies: %s", dep1)
	}

}

func makeRelease(r string, proj string) (release *qserver.Release, project *qproject.Project) {
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

func dfile1() []byte {
	data := `""" -*- coding: utf-8 -*-
About: API voor catalografische beschrijvingen.
"""
	
macro getCatGenStatus($data, $cloi):
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    '''
	«d %GetSs^gbcat(.$data,$cloi)»
	
macro setCatGenStatus($cloi, $staff, $mode, $time=""):
    '''
    $synopsis: Bewaart de statusvelden van een catalografische beschrijving
    $cloi: bibliografisch recordnummer in exchange format
    $staff: userid
    $mode: c: controle mode
           anders: editeer mode
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $example: m4_setCatGenStatus("c:lvd:1345679","rphilips","c",$h)
    '''
	«m4_CO d %SetSs^gbcat($cloi,$staff,$mode,$time)»`
	return []byte(data)
}

func dfile2() []byte {
	data := `""" -*- coding: utf-8 -*-
About: API voor catalografische beschrijvingen.
"""
	
macro getCatGenStatus($data, $cloi):
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $cloi: bibliografisch recordnummer in exchange format
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    '''
	«d %GetSs^gbcat(.$data,$cloi)»
	
macro setCatGenStatus($cloi, $staff, $mode, $time=""):
    '''
    $synopsis: Bewaart de statusvelden van een catalografische beschrijving
    $cloi: bibliografisch recordnummer in exchange format
    $staff: userid
    $mode: c: controle mode
           anders: editeer mode
    $time: Optioneel. Tijdstip laatste wijziging in $h formaat. default=$h.
    $example: m4_setCatGenStatus("c:lvd:1345679","rphilips","c",$h)
    '''
	«d %SetSs^gbcat($cloi,$staff,$mode,$time)»`
	return []byte(data)
}

func dfile3() []byte {
	data := `""" -*- coding: utf-8 -*-
About: Test
"""
	
m4_getCatGenStatus(.RAdata, "c:lvd:9999")

`
	return []byte(data)
}
