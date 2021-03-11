package bfile

import (
	"strings"
	"testing"

	qobject "brocade.be/qtechng/lib/object"
)

func TestParse01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello
	
loi c:
    $title: Catalografische beschrijving_loi
    $scope:
    $resolver: $$%Resolve^bbibresv
    $data:
    $md5global: ^ZZAD("md5")
    $md5index: ^ZZAD("md5i")
    $lookup: bibrec
    $genattr:
    $exist: s z1=$p(RDloi,":",2),z2=$p(RDloi,":",3,4),RDloiIs=$s(z2="":0,z1="":0,1:$d(^BCAT(z1,z2))'<10) // hello
    $next: s RDnxloi=$$%Nxt^uglcloi("^BCAT",RDloi)
    $previous: s RDpvloi=$$%Nxt^uglcloi("^BCAT",RDloi,-1)
	$inxinit: s RDssenv="" // Hello
//A
//B
    $inxsort: %GenSort^bcasix
    $inxgen: %GenIx^bcasix
    $inxrsv: « ^XBCAT »
    $genopac: d %GetOpac^gbcath(.RAopacs,RDigloi)
    $inxdata:
    $shortdesc: %Short^bcawshrt
    $fulldesc: bibrecpref
    $opacrofd: %Entry^bopwcloi
    $visual: %Visual^bcawvisl
    $consrsv: ⟦ ^BCATCONS ⟧
    $consexec: d %EntryC^bcawafim
    $metaformats: catxml; oai_dc; marc21; antilope; mods; umods; mmods; palsmarc21
    $ac2ac: d %Cloi^bcaschac(RDloi,RDacold,RDacnew)
    $globalexe: s RGa="^BCAT("_$p(RDloi,":",2) s:$p(RDloi,":",3)'="" RGa=RGa_","_$p(RDloi,":",3)
    $scanpat: c:
    $scanacc: abcdefghijklmnopqrstuvwxyz1234567890:
    $contgrf: ^BCAT(,)
    $contexe: s RDloi="c:"_PAst(1)_":"_PAst(2)
	$locktypes: «*,-p,-o
	





	hello»
    $lockglobal: ^XBLKDATA("c")
    $setexe: s RDset=$s(RDloi?1"c:"1.e1":"1.n:$p(RDloi,":",1,2),1:"") s:RDset'="" RAset(RDloi)=""
    $numeric: 1
    $mdexe: s RDh=$$%GetMD^gbcat(RDloi)
	$xmenu: d:3 %C^bcasmenu(.RAmenu,RDuser,RDinput)
	
	
	`)

	bfile := new(BFile)
	bfile.SetRelease("1.11")
	bfile.SetEditFile("hello/world")
	err := qobject.Loads(bfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	// Comment
	expect := `// -*- coding: utf-8 -*-
// About: Hello`
	if expect != bfile.Preamble {
		t.Errorf("Comment found: `%s`", bfile.Preamble)
	}
	// Number of brobs
	brobs := bfile.Brobs
	if len(brobs) != 1 {
		t.Errorf("# brobs found: `%d`", len(brobs))
		return
	}

	brob := brobs[0]

	// brob release
	if brob.Release() != "1.11" {
		t.Errorf("brob release is not set: %s", brob.Release())
		return
	}

	// brob editfile
	if brob.EditFile() != "hello/world" {
		t.Errorf("brob editfile is not set: %s", brob.EditFile())
		return
	}

	// brob id
	parts := strings.SplitN(brob.ID, " ", -1)
	id := parts[0]
	ty := parts[1]

	if id != "c" {
		t.Errorf("ID1: `%s`", brob.ID)
		return
	}

	// brob type

	if ty != "loi" {
		t.Errorf("Ty1: `%s`", ty)
		return
	}

	// specials

	m := brob.Map("", 1)

	key := "inxinit"
	evalue := `s RDssenv=""`
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	key = "inxrsv"
	evalue = ` ^XBCAT `
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	key = "consrsv"
	evalue = ` ^BCATCONS `
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	key = "locktypes"
	if !strings.HasSuffix(m[key], "hello") {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

}

func TestParse02(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello
	
mprocess %Search^pcasnoup:
    $title: searchmissingup
    $scope: Zoek, binnen een regelwerk, naar alle objecten zonder uitleenparameter / ongeldige uitleenparameter
    $cat: ABC
    $fp: PDcatsys
        $$scope: Catalografisch regelwerk
    $fp: PDlist1
        $$scope: Lijst met objecten zonder uitleenparemeter
    $fp: PDlist2
		$$scope: Lijst met objecten met ongeldige uitleenparameter
		
	
	`)

	bfile := new(BFile)
	bfile.SetRelease("1.11")
	bfile.SetEditFile("hello/world")
	err := qobject.Loads(bfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	// Comment
	expect := `// -*- coding: utf-8 -*-
// About: Hello`
	if expect != bfile.Preamble {
		t.Errorf("Comment found: `%s`", bfile.Preamble)
	}
	// Number of brobs
	brobs := bfile.Brobs
	if len(brobs) != 1 {
		t.Errorf("# brobs found: `%d`", len(brobs))
		return
	}

	brob := brobs[0]

	// brob release
	if brob.Release() != "1.11" {
		t.Errorf("brob release is not set: %s", brob.Release())
		return
	}

	// brob editfile
	if brob.EditFile() != "hello/world" {
		t.Errorf("brob editfile is not set: %s", brob.EditFile())
		return
	}

	// brob id

	parts := strings.SplitN(brob.ID, " ", -1)
	id := parts[0]
	ty := parts[1]

	if id != "%Search^pcasnoup" {
		t.Errorf("ID1: `%s`", id)
		return
	}

	// brob type

	if ty != "mprocess" {
		t.Errorf("Ty1: `%s`", ty)
		return
	}

	// specials

	m := brob.Map("", 1)

	key := "cat"
	evalue := `ABC`
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	key = "fp"
	evalue = `PDcatsys`
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	m = brob.Map("fp", 1)

	key = "scope"
	evalue = `Catalografisch regelwerk`
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

	m = brob.Map("fp", 2)

	key = "scope"
	evalue = `Lijst met objecten zonder uitleenparemeter`
	if m[key] != evalue {
		t.Errorf("%s: `%s`", key, m[key])
		return
	}

}
