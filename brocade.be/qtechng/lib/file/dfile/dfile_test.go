package dfile

import (
	"testing"
	qobject "brocade.be/qtechng/lib/object"
)

func TestParse01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello

macro getCatGenStatus($data, *$cloi=(A+B")")):
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
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»
	
	
	`)

	dfile := new(DFile)
	dfile.SetRelease("1.11")
	dfile.SetEditFile("hello/world")
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	// Comment
	expect := `// -*- coding: utf-8 -*-
// About: Hello`
	if expect != dfile.Preamble {
		t.Errorf("Comment found: `%s`", dfile.Preamble)
	}
	// Number of macros
	macros := dfile.Macros
	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	macro := macros[0]

	// macro release
	if macro.Release() != "1.11" {
		t.Errorf("macro release is not set: %s", macro.Release())
		return
	}

	// macro editfile
	if macro.EditFile() != "hello/world" {
		t.Errorf("macro editfile is not set: %s", macro.EditFile())
		return
	}

	// macro id

	if macro.ID != "getCatGenStatus" {
		t.Errorf("ID1: `%s`", macro.ID)
		return
	}

	// macro synopsis

	if macro.Synopsis != "Bepaalt het statusveld bij een bibliografische beschrijving" {
		t.Errorf("Synopsis1: `%s`", macro.Synopsis)
		return
	}

	// macro examples

	if len(macro.Examples) != 2 {
		t.Errorf("# examples found: `%d`", len(macro.Examples))
		return
	}
	example1 := macro.Examples[0]
	if example1 != `m4_getCatGenStatus(Array,"c:lvd:1345679")` {
		t.Errorf("Example1: `%s`", example1)
		return
	}
	example2 := macro.Examples[1]
	if example2 != `m4_getCatGenStatus(Array,"c:lvd:13hh5679")` {
		t.Errorf("Example2: `%s`", example2)
		return
	}

	// macro parameters

	if len(macro.Params) != 2 {
		t.Errorf("Number of  paramas should be 2")
		return
	}

	param1 := macro.Params[0]
	param2 := macro.Params[1]

	if param1.ID != "$data" {
		t.Errorf("Name1: `%s`", param1.ID)
		return
	}
	if param2.ID != "$cloi" {
		t.Errorf("Name2: `%s`", param1.ID)
		return
	}

	if param1.Named {
		t.Errorf("Name1: should not be named")
		return
	}
	if !param2.Named {
		t.Errorf("Name2: should be named")
		return
	}

	if param1.Default != `` {
		t.Errorf("Default1: `%s`", param2.Default)
		return
	}

	if param2.Default != `(A+B")")` {
		t.Errorf("Default2: `%s`", param2.Default)
		return
	}

	// Actions

	if len(macro.Actions) != 2 {
		t.Errorf("Number of  actions is `%d`: `%v`", len(macro.Actions), macro.Actions)
		return
	}
}

func TestParse02(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello

macro getCatGenStatus($data, *$cloi=(A+B")")):
    '''
    $synopsis: xxx
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	macro setCatGenStatus($data, *$cloi=(A+B")")):
    '''
    $synopsis: xxx
    $cloi: Hello World
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»
	`)

	dfile := new(DFile)
	dfile.SetRelease("1.11")
	dfile.SetEditFile("hello/world")
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if dfile.Macros[1].Params[0].Doc != dfile.Macros[0].Params[0].Doc {
		t.Errorf("Error: \n%s\n%s\n", dfile.Macros[1].Params[0].Doc, dfile.Macros[0].Params[0].Doc)
		return
	}
	if dfile.Macros[1].Params[1].Doc == dfile.Macros[1].Params[0].Doc {
		t.Errorf("Error: \n%s\n%s\n", dfile.Macros[1].Params[0].Doc, dfile.Macros[0].Params[0].Doc)
		return
	}

	if dfile.Macros[0].Lint() != nil {
		t.Errorf("Error: \n%s\n", dfile.Macros[0].Lint())
		return
	}

	if dfile.Macros[1].Lint() != nil {
		t.Errorf("Error: \n%s\n", dfile.Macros[1].Lint())
		return
	}

}

func TestParam01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello

macro getCatGenStatus(    ):
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving

    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 0 {
			t.Errorf("Number of  paramas should be 0")
		}
	}
}

func TestParam02(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello

macro getCatGenStatus:
    '''
    $synopsis: Bepaalt het statusveld bij een bibliografische beschrijving

    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 0 {
			t.Errorf("Number of  paramas should be 0")
		}
	}
}

func TestParam03(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello

macro getCatGenStatus(   $data , $b          ,    *$c):
    '''
    $synopsis: xxx
    $data: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
	$b: bibliografisch recordnummer in exchange format
	$c: Hello
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)

	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 3 {
			t.Errorf("Number of  paramas should be 3")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$data" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[1].ID != "$b" {
		t.Errorf("Error: param[1] name is %s", params[1].ID)
	}
	if params[2].ID != "$c" {
		t.Errorf("Error: param[2] name is %s", params[2].ID)
	}
	if params[0].Named {
		t.Errorf("Error: param[0] should not be named")
	}
	if params[1].Named {
		t.Errorf("Error: param[1] should not be named")
	}
	if !params[2].Named {
		t.Errorf("Error: param[2] should be named")
	}

	if params[0].Default != "" {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
	if params[1].Default != "" {
		t.Errorf("Error: param[1] has default `%s`", params[1].Default)
	}
	if params[2].Default != "" {
		t.Errorf("Error: param[2] has default `%s`", params[2].Default)
	}

}

func TestParam04(t *testing.T) {
	preamble := "// About\n"
	data := []byte(preamble + `macro Third(   $a   = «»,   $b  =  « a »   , *$c =  «»)   :
	'''
    $synopsis: xxx
    $a: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
	$b: bibliografisch recordnummer in exchange format
	$c: Hello
    $example: m4_getCatGenStatus(Array,"c:lvd:1345679")
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 3 {
			t.Errorf("Number of  paramas should be 3")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$a" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[1].ID != "$b" {
		t.Errorf("Error: param[1] name is %s", params[1].ID)
	}
	if params[2].ID != "$c" {
		t.Errorf("Error: param[2] name is %s", params[2].ID)
	}
	if params[0].Named {
		t.Errorf("Error: param[0] should not be named")
	}
	if params[1].Named {
		t.Errorf("Error: param[1] should not be named")
	}
	if !params[2].Named {
		t.Errorf("Error: param[2] should be named")
	}

	if params[0].Default != "" {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
	if params[1].Default != " a " {
		t.Errorf("Error: param[1] has default `%s`", params[1].Default)
	}
	if params[2].Default != "" {
		t.Errorf("Error: param[2] has default `%s`", params[2].Default)
	}
}

func TestParam05(t *testing.T) {
	preamble := "// About\n"
	data := []byte(preamble + `macro Third(   $1   = abc )   :
'''
    $synopsis: xxx
    $1: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 1 {
			t.Errorf("Number of  paramas should be 1")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$1" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[0].Default != "abc" {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
}

func TestParam06(t *testing.T) {
	preamble := "// About\n"
	data := []byte(preamble + `macro Third(   $1   = 4 - "a,b)c"+5 )   :
	'''
    $synopsis: xxx
    $1: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 1 {
			t.Errorf("Number of  paramas should be 1")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$1" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[0].Default != `4 - "a,b)c"+5` {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
}

func TestParam07(t *testing.T) {
	preamble := "// About\n"
	data := []byte(preamble + `macro Third(   $1   = (A,B) )  :
	'''
    $synopsis: xxx
    $1: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 1 {
			t.Errorf("Number of  paramas should be 1")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$1" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[0].Default != `(A,B)` {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
}

func TestParam08(t *testing.T) {
	preamble := "// About\n"
	data := []byte(preamble + `macro Third(   $1   = (A,B ")" ) )  :
	'''
    $synopsis: xxx
    $1: Array die de statusvelden bevat.
            ss: de inhoud van het statusveld (edittoken)
            cp: creator id van de record
            cd: creatie tijdstip van dit record
            mp: id van de persoon die deze record laatst gewijzigd heeft
            md: tijdstip waarop deze record laatst gewijzigd werd
            tp: id van de persoon die deze record laatst gecontroleerd heeft
            td: tijdstip waarop deze record laatst gecontroleerd werd
            st: de status van het record (d=deleted)
    $example: m4_getCatGenStatus(Array,"c:lvd:13hh5679")
    '''
	«d %GetSs^gbcat(.$data, $cloi)» if «$cloi isInstanceOf "numlit"»
	«d %GetSs^gbcat(.$data, 1+$cloi)»
	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	if len(macros) == 1 {
		macro := macros[0]
		if len(macro.Params) != 1 {
			t.Errorf("Number of  paramas should be 1")
			return
		}
	}

	params := macros[0].Params
	if params[0].ID != "$1" {
		t.Errorf("Error: param[0] name is %s", params[0].ID)
	}
	if params[0].Default != `(A,B ")" )` {
		t.Errorf("Error: param[0] has default `%s`", params[0].Default)
	}
}

func TestParse11(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
	macro Third ($cloi, $eloi) :

	'''
	$synopsis: Uitleg1T
	$cloi: C-loi
	$eloi: E-loi
    $example: m4_Third()
'''
$cloi, $eloi

	`)
	dfile := new(DFile)
	err := qobject.Loads(dfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	macros := dfile.Macros

	if len(macros) != 1 {
		t.Errorf("# macros found: `%d`", len(macros))
		return
	}

	macro := macros[0]
	if macro.Synopsis != "Uitleg1T" {
		t.Errorf("Synopsis11: `%s`", macro.Synopsis)
	}
	if len(macro.Params) != 2 {
		t.Errorf("# params found: `%d`", len(macro.Params))
		return
	}
	param := macro.Params[1]
	if param.ID != "$eloi" {
		t.Errorf("param2 Name is: `%s`", param.ID)
		return
	}
	if param.Doc != "E-loi" {
		t.Errorf("param2 doc is: `%s`", param.Doc)
		return
	}
	examples := macro.Examples

	if len(examples) != 1 {
		t.Errorf("# examples is: `%d`", len(examples))
		return
	}
	example := examples[0]
	if example != "m4_Third()" {
		t.Errorf("example is: `%s`", example)
	}

}
