package dfile

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"testing"

	qregistry "brocade.be/base/registry"
	qobject "brocade.be/qtechng/object"
	qserver "brocade.be/qtechng/server"
	qutil "brocade.be/qtechng/util"
)

func TestMacroFile01(t *testing.T) {

	s := `// About: macro parsing
	macro    ABC   :   
	'''
	$synopsis: Hallo
			   ABC
	$example: m4_ABC
	'''
		abc
		`

	blob := []byte(s)

	dfile := new(DFile)
	err := qobject.Loads(dfile, blob)

	if err != nil {
		t.Errorf("Parse error:\n%s", err)
		return
	}

	if len(dfile.Objects()) != 1 {
		t.Errorf("Aantal macro's zou 1 moeten zijn: %d", len(dfile.Objects()))
		return
	}
	macros := dfile.Objects()
	macro := macros[0].(*Macro)

	if macro.Name() != "ABC" {
		t.Errorf("Naam van macro zou ABC moeten zijn")
		return
	}

	// if macro.Name() != "ABC" {
	// 	t.Errorf("Name error: found `%s`\n", macro.Name())
	// 	return
	// }
	// if len(macro.Params) != 0 {
	// 	t.Errorf("Param error: found `%d` params\n", len(macro.Params))
	// 	return
	// }
}

func TestMacroFile02(t *testing.T) {

	testfile := path.Join(qregistry.Registry["qtechng-test-dir"], "cat.d")

	blob, _ := ioutil.ReadFile(testfile)

	macrofile := new(DFile)
	err := macrofile.Loads("", blob, "9.99")

	if err != nil {
		t.Errorf("Parse error:\n%s", err)
		return
	}

	if len(macrofile.Objects()) != 168 {
		t.Errorf("Aantal macro's zou 1 moeten zijn, gevonden: %d", len(macrofile.Objects()))
		return
	}
	macros := macrofile.Objects()
	macro := macros[0].(*Macro)

	if macro.Name() != "getCatGenStatus" {
		t.Errorf("Naam van macro zou getCatGenStatus moeten zijn")
		return
	}

}

func TestMacro01(t *testing.T) {

	s := `macro    ABC   :   
	'''
	$synopsis: Hallo
			   ABC
	$example: m4_ABC
	'''
		abc
		`

	blob := []byte(s)

	macro := new(Macro)
	err := qobject.Parse(macro, "", blob, "9.99")

	if err != nil {
		t.Errorf("Parse error:\n%s", err)
		return
	}

	if macro.Name() != "ABC" {
		t.Errorf("Name error: found `%s`\n", macro.Name())
		return
	}
	if len(macro.Params) != 0 {
		t.Errorf("Param error: found `%d` params\n", len(macro.Params))
		return
	}
}

func TestMacro02(t *testing.T) {

	s := `macro ABC (   ) :
	'''
	$synopsis: Hallo
			   ABC
	$example: m4_ABC
	'''
		abc
		`

	blob := []byte(s)

	macro := new(Macro)
	err := qobject.Parse(macro, "", blob, "9.99")

	if err != nil {
		t.Errorf("Parse error:\n%s", err)
		return
	}

	if macro.Name() != "ABC" {
		t.Errorf("Name error: found `%s`\n", macro.Name())
		return
	}
	if len(macro.Params) != 0 {
		t.Errorf("Param error: found `%d` params\n", len(macro.Params))
		return
	}
}

func TestMacro03(t *testing.T) {

	s := `macro ABC($a , *$bb = 27, $ac = "h" ):
	'''
	$synopsis: Hallo
			   ABC
	$a: A-info
	$bb: B-info1
	    B-info2
	$example: m4_ABC(1, 2)
	'''
		$ac + $a + $bb + $ac
		`

	blob := []byte(s)

	macro := new(Macro)
	err := qobject.Parse(macro, "", blob, "9.99")

	if err != nil {
		t.Errorf("Parse error:\n%s", err)
		return
	}

	if macro.Name() != "ABC" {
		t.Errorf("Name error: found `%s`\n", macro.Name())
		return
	}

	if len(macro.Params) != 3 {
		t.Errorf("Param error: found `%d` params\n", len(macro.Params))
		return
	}

	if macro.Params[0].ID != "$a" {
		t.Errorf("Param 1 error: found name `%s`\n", macro.Params[0].ID)
		return
	}

	if macro.Params[0].Default != "" {
		t.Errorf("Param 1 error: found default `%s`\n", macro.Params[0].Default)
		return
	}

	if macro.Params[0].Named {
		t.Errorf("Param 1 error: found named `%v`\n", macro.Params[0].Named)
		return
	}

	if macro.Params[1].ID != "$bb" {
		t.Errorf("Param 2 error: found name `%s`\n", macro.Params[1].ID)
		return
	}

	if macro.Params[1].Default != "27" {
		t.Errorf("Param 2 error: found default `%s`\n", macro.Params[1].Default)
		return
	}

	if !macro.Params[1].Named {
		t.Errorf("Param 2 error: found named `%v`\n", macro.Params[1].Named)
		return
	}

	if macro.Params[2].ID != "$ac" {
		t.Errorf("Param 3 error: found name `%s`\n", macro.Params[2].ID)
		return
	}

	if macro.Params[2].Default != `"h"` {
		t.Errorf("Param 3 error: found default `%s`\n", macro.Params[2].Default)
		return
	}

	if !macro.Params[2].Named {
		t.Errorf("Param 2 error: found named `%v`\n", macro.Params[2].Named)
		return
	}

	if len(macro.Actions) != 1 {
		t.Errorf("Action error: found `%d` actions\n", len(macro.Actions))
		return
	}

	if len(macro.Actions) != 1 {
		t.Errorf("Action error: found `%d` actions\n", len(macro.Actions))
		return
	}

	binary := macro.Actions[0].Binary
	if !qutil.CmpStringSlice(binary, []string{"", "$ac", " + ", "$a", " + ", "$bb", " + ", "$ac", ""}, false) {
		t.Errorf("Action binary error: found `%v` \n", binary)
		return
	}

	env := map[string]string{
		"$a":  "AA",
		"$ac": "AC",
	}

	if macro.Replacer(env, "Hallo") != "AC + AA + 27 + AC" {
		t.Errorf("Replacer error: found `%s` \n", macro.Replacer(env, "Hallo"))
		return
	}

}

func TestMacroParam01(t *testing.T) {

	type T struct {
		signature string
		names     []string
		defaults  []string
	}

	test := []T{
		{
			"Call()",
			[]string{},
			[]string{},
		},
		{
			"Call($1)",
			[]string{"$1"},
			[]string{""},
		},
		{
			"Call($1)",
			[]string{"$1"},
			[]string{""},
		},
		{
			"Call($a1)",
			[]string{"$a1"},
			[]string{""},
		},
		{
			"Call($1,$a1)",
			[]string{"$1", "$a1"},
			[]string{"", ""},
		},
		{
			`Call($a= 27, $b =" aaaa "   )`,
			[]string{"$a", "$b"},
			[]string{"27", `" aaaa "`},
		},

		{
			"Call($a=A(B(C)))",
			[]string{"$a"},
			[]string{"A(B(C))"},
		},
		{
			"Call($a=A(B(C,$b)))",
			[]string{"$a"},
			[]string{"A(B(C,$b))"},
		},

		{
			`Call(  $a   =        A(B(C,$b,")")))`,
			[]string{"$a"},
			[]string{`A(B(C,$b,")"))`},
		},

		{
			`Call(  $a   =« ) »   )`,
			[]string{"$a"},
			[]string{` ) `},
		},

		{
			`Call(  $a   = ⟦ » ⟧    )`,
			[]string{"$a"},
			[]string{` » `},
		},
	}

	for _, tst := range test {
		s := tst.signature[5:]
		x, _ := Parse("", []byte(s), Entrypoint("Params"))
		params := make([]Param, 0)
		if x != nil {
			params = x.([]Param)
		}

		if len(params) != len(tst.names) {
			t.Errorf("Param for `%v`: found `%d`", tst, len(params))
			return
		}
		if len(params) == 0 {
			continue
		}
		for i, param := range params {
			if param.ID != tst.names[i] {
				t.Errorf("Param name for `%v`: found `%s`", tst, param.ID)
				return
			}
			if param.Default != tst.defaults[i] {
				t.Errorf("Param default for `%v`: found `%s`", tst, param.Default)
				return
			}

		}

	}

}

func TestMacro011(t *testing.T) {

	macro := Macro{
		ID:     "macroName",
		Source: "/a/b/c.d",
	}

	if name := macro.Name(); name != "macroName" {
		t.Errorf("Macro name: found `%s`", name)
		return
	}

	if ty := macro.Type(); ty != "m4" {
		t.Errorf("Macro : found `%s`", ty)
		return
	}

	if editfile := macro.EditFile(); editfile != "/a/b/c.d" {
		t.Errorf("Macro : found `%s`", editfile)
		return
	}

	slob := `{
	"id": "macroName",
	"synopsis": "",
	"params": null,
	"actions": null,
	"examples": null,
	"source": "/a/b/c.d",
	"lineno": "",
	"release": ""
}`
	trimmer := func(s string) string {
		space := regexp.MustCompile(`\s+`)
		s = space.ReplaceAllString(s, "")
		return s
	}

	json, _ := macro.Marshal()
	if trimmer(string(json)) != trimmer(slob) {
		t.Errorf("Macro : found \n`%s`\n\nshould have been:\n`%s`\n", string(json), slob)
		return
	}

	json = []byte(`{"id":"macroName2","synopsis":"","params":null,"actions":null,"examples":null,"source":"/a/b/c.d","lineno":"","release":""}`)

	_ = macro.Unmarshal(json)

	if name := macro.Name(); name != "macroName2" {
		t.Errorf("Macro name: found `%s`", name)
		return
	}

}

func TestLoads01(t *testing.T) {

	macro := new(Macro)
	macro.Source = "/a/b/c.d"

	body := []byte(`macro getCatGenStatus2($data, *$cloi=(A+B")")):
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
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data,$cloi)» if «true»
	«d %GetAB^gbcat(.$data,$cloi)» unless «false»
	«d %GetSs^gbcat(.$data,$cloi)» if «$cloi isInstanceOf "numlit"»
	«d %ZZZGetSs^gbcat(.$data, 1+$cloi)»`)

	err := qobject.Parse(macro, "", body, "9.99")

	if err != nil {
		t.Errorf("Error:\n%s", err)
	}
	if name := macro.Name(); name != "getCatGenStatus2" {
		t.Errorf("Macro name: found `%s`", name)
		return
	}

	if edit := macro.EditFile(); edit != "/a/b/c.d" {
		t.Errorf("Macro source: found `%s`", edit)
		return
	}

	x, _ := macro.Marshal()

	f := macro.Format()

	macro2 := new(Macro)
	macro2.Source = "/a/b/c.d"

	err2 := qobject.Parse(macro2, "", []byte(f), "9.99")

	if err2 != nil {
		t.Errorf("Error2:\n%s", err2)
	}
	f2 := macro2.Format()

	if f != f2 {
		t.Errorf("Macro source: found error in formats")
		fmt.Println(string(x))
		fmt.Println(f)
		fmt.Println(f2)
		return
	}

	release, _ := qserver.Release{}.New("1.94", false)
	release.FS().RemoveAll("/")

	err = release.Init()
	if err != nil {
		t.Errorf("Creation failed `%s`", err)
		return
	}

	macro.SetRelease("1.94")
	change1, e1 := qobject.Store(macro)
	if e1 != nil {
		t.Errorf("Store failed `%s`", e1)
		return
	}

	if !change1 {
		t.Errorf("Should changed")
		return
	}

	m2 := new(Macro)
	m2.SetRelease(macro.Release())
	m2.SetName(macro.Name())
	er := qobject.Fetch(m2)

	if er != nil {
		t.Errorf("fetch failed `%s`", er)
		return
	}

	if macro.Synopsis != m2.Synopsis {
		t.Errorf("fetch failed !")
		return
	}

	change2, e2 := qobject.Store(macro)
	if e2 != nil {
		t.Errorf("Store failed `%s`", e2)
		return
	}

	if change2 {
		t.Errorf("Should NOT changed")
		return
	}
	change3, e3 := qobject.Waste(macro)
	if e3 != nil {
		t.Errorf("Waste failed `%s`", e3)
		return
	}

	if !change3 {
		t.Errorf("Should be changed")
		return
	}

	change4, e4 := qobject.Waste(macro)
	if e4 != nil {
		t.Errorf("Waste failed `%s`", e4)
		return
	}

	if !change4 {
		t.Errorf("Should NOT be changed")
		return
	}

}

func TestLoads02(t *testing.T) {

	macro := new(Macro)
	macro.Source = "/a/b/c.d"

	body := []byte(`macro getCatGenStatus($dat, *$cloi=(A+B")")):
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving
    $dat: Array die de statusvelden bevat.
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
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data,$cloi)» if «true»
	«d %GetAB^gbcat(.$data,$cloi)» unless «false»
	«d %GetSs^gbcat(.$data,$cloi)» if «$cloi isInstanceOf "numlit"»
	«d %ZZZGetSs^gbcat(.$dat, 1+$cloi)»`)

	err := qobject.Parse(macro, "", body, "1.94")

	if err != nil {
		t.Errorf("Error:\n%s", err)
	}

	release, _ := qserver.Release{}.New("1.94", false)
	release.FS().RemoveAll("/")

	err = release.Init()
	if err != nil {
		t.Errorf("Creation failed `%s`", err)
		return
	}

	macro.SetRelease("1.94")

	macrolist := make([]interface{}, 0)

	name := macro.Name()
	for i := 0; i < 10000; i++ {
		m := new(Macro)
		*m = *macro
		m.SetName(name + strconv.Itoa(i))
		macrolist = append(macrolist, m)
		// m.Store()
	}
	changedlist, _ := qobject.StoreList(macrolist)

	count := 0

	for _, c := range changedlist {
		if c {
			count++
		}
	}

	if count != len(macrolist) {
		t.Errorf("Creation failed: only %d changed", count)
		return
	}

	changedlist, _ = qobject.StoreList(macrolist)

	count = 0

	for _, c := range changedlist {
		if c {
			count++
		}
	}

	if count != 0 {
		t.Errorf("Creation failed: only %d changed", count)
		return
	}

}
