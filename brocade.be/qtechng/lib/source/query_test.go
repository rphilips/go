package source

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	qmeta "brocade.be/qtechng/lib/meta"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
)

func TestQuery01(t *testing.T) {
	r := "9.99"
	makeqRelease(r)
	p := "/a/b/c/d/f1.txt"

	source, err := Source{}.New(r, p, false)
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	// release
	query := &Query{
		Release: "9.99",
	}

	result := source.Test(query)

	if !result {
		t.Errorf("Should match !")
		return
	}

	query = &Query{
		Release: "9.98",
	}

	result = source.Test(query)

	if result {
		t.Errorf("Should not match !")
		return
	}

	// patterns
	query = &Query{
		Release:  "9.99",
		Patterns: []string{"/a/*"},
	}

	result = source.Test(query)

	if !result {
		t.Errorf("Should match !")
		return
	}
	query = &Query{
		Release:  "9.99",
		Patterns: []string{"/a/d*"},
	}

	result = source.Test(query)

	if result {
		t.Errorf("Should not match !")
		return
	}

	query = &Query{
		Release:  "9.98",
		Patterns: []string{"/a/*"},
	}

	result = source.Test(query)

	if result {
		t.Errorf("Should not match !")
		return
	}

	query = &Query{
		Release:  "9.99",
		Patterns: []string{"/a/d*", "*txt"},
	}

	result = source.Test(query)

	if !result {
		t.Errorf("Should match !")
		return
	}

	query = &Query{
		Release:  "9.99",
		Patterns: []string{"/a/d*", "*[t]xt"},
	}

	result = source.Test(query)

	if !result {
		t.Errorf("Should match !")
		return
	}
}
func TestQuery02(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release: r,
	}

	result := query.Run()

	if len(result) != 12 {
		t.Errorf("Should have found 12: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  "9.99",
		Contains: []string{"Hello"},
	}

	result = query.Run()

	if len(result) != 9 {
		t.Errorf("Hello Should have found 9: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  "9.99",
		Contains: []string{"pdf"},
	}

	result = query.Run()

	if len(result) != 3 {
		t.Errorf("Hello Should have found 3: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func TestQuery028(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release:  "9.99",
		Contains: []string{"schema"},
	}

	result := query.Run()

	if len(result) != 3 {
		t.Errorf("Hello Should have found 3: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}
func TestQuery12(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release:  r,
		Patterns: []string{"/a"},
	}

	result := query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 8: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  r,
		Patterns: []string{"/ab", "/a"},
	}

	result = query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 8: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  r,
		Patterns: []string{"/a/b.xyz", "/a"},
	}

	result = query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 8: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  r,
		Patterns: []string{"/a/b.*xyz", "/a"},
	}

	result = query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 8: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  r,
		Patterns: []string{"/a/b.*xyz", "/a"},
	}

	result = query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 8: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func BenchmarkQuery77(t *testing.B) {
	r := "9.98"

	query := &Query{
		Release:  r,
		Patterns: []string{"/*"},
		Contains: []string{"m4_getCatIsbdTitles"},
	}

	result := query.Run()

	if len(result) != 74 {
		t.Errorf("Should have found 74: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}
}

func BenchmarkQuery78(t *testing.B) {
	r := "9.98"

	query := &Query{
		Release:  r,
		Patterns: []string{"/*.m"},
		Contains: []string{"m4_getCatIsbdTitles"},
	}

	result := query.Run()

	if len(result) != 72 {
		t.Errorf("Should have found 72: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func BenchmarkQuery79(t *testing.B) {
	r := "9.98"

	query := &Query{
		Release:  r,
		Patterns: []string{"/*"},
		Contains: []string{"m4_getCatIsbdTitles"},
		All: [](func(qpath string, blob []byte) bool){
			func(qpath string, blob []byte) bool {
				k := strings.LastIndex(qpath, "/")
				if k == -1 {
					return false
				}
				base := qpath[k+1:]
				if len(base) < 4 {
					return false
				}
				fourth := rune(base[3])
				ok := strings.HasPrefix(base, "g") || strings.ContainsRune("wuts", fourth)
				return !ok
			},
		},
	}

	result := query.Run()

	if len(result) != 13 {
		t.Errorf("Should have found 72: %d", len(result))
		fmt.Println(query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func TestQuery025(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release:  "9.99",
		Contains: []string{"Hel+o"},
		Regexp:   true,
	}

	result := query.Run()

	if len(result) != 9 {
		t.Errorf("Hello Should have found 9: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func TestQuery026(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release:  "9.99",
		Contains: []string{"Hel+o.*\n.*Moon"},
		Regexp:   true,
		PerLine:  false,
	}

	result := query.Run()

	if len(result) != 9 {
		t.Errorf("Hello(1) Should have found 9: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  "9.99",
		Contains: []string{"Hello", "World"},
		PerLine:  true,
	}

	result = query.Run()

	if len(result) != 9 {
		t.Errorf("Hello(2) Should have found 9: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  "9.99",
		Contains: []string{"Hello", "World"},
		PerLine:  true,
		QDirs:    []string{"/a"},
	}

	result = query.Run()

	if len(result) != 6 {
		t.Errorf("Hello(3) Should have found 6: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release:  "9.99",
		Contains: []string{"Hello", "Moon"},
		PerLine:  true,
	}

	result = query.Run()

	if len(result) != 0 {
		t.Errorf("Hello Should have found 9: %d", len(result))
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func TestQuery03(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release: r,
		Natures: []string{"config"},
	}

	result := query.Run()

	if len(result) != 3 {
		t.Errorf("Should have found 3: %d", len(result))
		return
	}

}

func TestQuery04(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	query := &Query{
		Release:  r,
		MtBefore: "2019-12-18",
	}

	result := query.Run()

	if len(result) != 6 {
		t.Errorf("Should have found 3: %d", len(result))
		for _, x := range result {
			meta, _ := qmeta.Meta{}.New(r, x.String())
			fmt.Println(x.String(), meta)
		}
		return
	}

}

func TestQuery05(t *testing.T) {
	r := "9.99"
	makeqRelease(r)

	fall1 := func(p string, blob []byte) bool {
		ok := len(p) < 15
		fmt.Println("fall1", p, ok)
		return ok
	}

	fall2 := func(p string, blob []byte) bool {
		ok := strings.Contains(p, "1")
		fmt.Println("fall2", p, ok)
		return ok
	}
	query := &Query{
		Release: r,
		All:     []func(string, []byte) bool{fall1, fall2},
	}

	result := query.Run()

	if len(result) != 4 {
		t.Errorf("Should have found 3: %d", len(result))
		fmt.Println("query:", query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

	query = &Query{
		Release: r,
		Any:     []func(string, []byte) bool{fall1, fall2},
	}

	result = query.Run()

	if len(result) != 8 {
		t.Errorf("Should have found 3: %d", len(result))
		fmt.Println("query:", query)
		for _, x := range result {
			fmt.Println(x.String())
		}
		return
	}

}

func makeqRelease(r string) (release *qserver.Release, project *qproject.Project) {
	release, _ = qserver.Release{}.New(r, false)
	release.FS("/").RemoveAll("/")
	err := release.Init()
	if err != nil {
		fmt.Println("error server:", err)
	}
	for _, proj := range []string{"/a/b/c/d", "/a1/b1", "/a/b"} {
		prj, _ := qproject.Project{}.New(r, proj, false)
		prj.Init(qmeta.Meta{})
		for i, fname := range []string{"f1.txt", "f2.bin", "f3.pdf"} {
			p := proj + "/" + fname
			source, _ := Source{}.New(r, p, false)
			meta := qmeta.Meta{}
			meta.Ct = "2019-02-18T12:00:00"
			meta.Mt = "2019-1" + strconv.Itoa(i) + "-18T13:00:00"
			meta.Cu = "mjeuris"
			meta.Mu = "rphilips"
			_, _, _, err = source.Store(meta, "Hello World "+fname+"\nBye Moon", false)
			if err != nil {
				fmt.Println("err:", err.Error())
			}
		}
	}
	return
}
