package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	s := `{"a": "A", "b": "B"}`
	var any interface{}
	e := json.Unmarshal([]byte(s), &any)
	if e != nil {
		t.Errorf("Error1: `%s`", e.Error())
	}
	_, e = json.MarshalIndent(any, "", "   ")
	if e != nil {
		t.Errorf("Error2: `%s`", e.Error())
	}
}

func TestTime01(t *testing.T) {
	x := Time("2020 . 07 . 08")
	y := Timestamp(true)
	if x != "2020-07-08T00:00:00" || y != "" {
		t.Errorf("Error: `%s`", x)
		t.Errorf("Errory: `%s`", y)
	}
}

func TestInfo01(t *testing.T) {
	data := "    $Synopsis   : Lijn 1 \n\n"

	r := Info(data, "$synopsis")

	if r != "Lijn 1" {
		t.Errorf("Error: `%s`", r)
	}
}

func TestInfo11(t *testing.T) {
	data := `    $synopsis:
Lijn 1

`

	r := Info(data, "$synopsis")

	if r != "Lijn 1" {
		t.Errorf("Error: `%s`", r)
	}
}

func TestInfo02(t *testing.T) {
	data := `    $Synopsis
: Lijn 1

`

	r := Info(data, "$synopsis")

	if r != "Lijn 1" {
		t.Errorf("Error: `%s`", r)
	}
}

func TestBM01(t *testing.T) {
	data := `// -*- coding: utf-8 -*-
	// About: Copieren van een catalografisch beschrijving


	def %Copy(PDcloi, PDreg, PAopt):
	 n x,y,z,return,RDloi,ins,RAoptc,RAispat,lm,archive,UDcaCode,error,user,RDcloi
	 s return=""
	 i $G(PDreg)="" s PDreg=$P(PDcloi,":",2)
	 s RDloi=$G(PAopt("target"))
	 i RDloi="" m4_newCatGenRecord(RDloi,PDreg)
	 ;default waarden
	 ;
	 i $g(PAopt("lmrule")) d
	 . k RAoptc
	 . m4_getCatMembershipPrimary(lm,PDcloi)
	 . m4_getCatLmCopyOptions(RAoptc,lm)
	 . f x="if","in","index","pk","relation","subject","access","lm" d
	 .. s PAopt(x)=$g(RAoptc(x))
	 .. q
	 . k PAopt("nr")
	 . m PAopt("nr")=RAoptc("nr")
	 . q
	 k RAispat
	 i $d(PAopt("user")) d
	 . s user=$g(PAopt("user")) s:user="" user=$g(UDuser)
	 . m4_getUserCatInstitutes(RAispat,user,catsys=$p(PDcloi,":",2))
	 . q
	 s error=0
	 i $d(PAopt("user")) d  q:error'=0 ""
	 . s user=$g(PAopt("user")) s:user="" user=$g(UDuser) q:user=""
	 . m4_getCatPermissionCopy(error,PDcloi,user)
	 . q
	 s:'$d(PAopt("subject")) PAopt("subject")=1
	 s:'$d(PAopt("index")) PAopt("index")=1
	 s return=RDloi
	 d LM(PDcloi,RDloi)
	 d DR(PDcloi,RDloi)
	 d LG(PDcloi,RDloi)
	 d TI(PDcloi,RDloi)
	 d AU(PDcloi,RDloi)
	 d CA(PDcloi,RDloi)
	 d ED(PDcloi,RDloi)
	 d IM(PDcloi,RDloi)
	 d CO(PDcloi,RDloi)
	 d NT(PDcloi,RDloi)
	 i $g(PAopt("if")) d IF(PDcloi,RDloi)
	 i $g(PAopt("in")) d IN(PDcloi,RDloi)
	 s tst=1
	 ;bij nr hoort wel degelijk een $d : switches staan op onderliggend niveau
	 i $d(PAopt("nr")) d NR(PDcloi,RDloi,.PAopt)
	 i $g(PAopt("subject")) d SU(PDcloi,RDloi)
	 s tst=1
	 i $g(PAopt("pk")) d
	 . i $d(PAopt("pk"))=1 d cpHolAll(PDcloi,RDloi,.RAispat)
	 . i $d(PAopt("pk"))>10 d
	 .. s ins=""
	 .. f  s ins=$O(PAopt("pk",ins)) q:ins=""  d
	 ... d cpHolIs(PDcloi,RDloi,ins,.RAispat)
	 ... q
	 .. q
	 . q
	 m4_finaliseCatRecord(RDloi)
	 m4_setCatOpacsStandalone(RDloi)
	 i $g(PAopt("relation")) d RL(PDcloi,RDloi)
	 ;index aanpassing moet altijd achteraan
	 i $g(PAopt("index")) d
	 . m4_updIndexCatLoi(RDloi)
	 . q
	 s UDcaCode=PDcloi
	 s archive=$G(PAopt("archive")) i archive="" s archive=1
	 i archive d
	 . n UDcaCode
	 . s (RDcloi,UDcaCode)=RDloi
	 . d %Archive^bcawfile
	 . m4_setCatGenStatus(RDloi,UDuser,"m",$h)
	 . q
	 m4_browseCatList(RDloi,UDuser)
	 m4_newCatList(RDloi,UDuser)
	 m4_userCatList(RDloi,UDuser)
	 q return
	 ;

	def LM(PDsource, PDtarget):
	 n x,y,z
	 n RAlm,lm,ZAoptc
	 m4_getCatIsbdMemberships(RAlm,PDsource)
	 s lm=""
	 f  s lm=$O(RAlm(lm)) q:lm=""  d
	 . m4_getCatLmCopyOptions(ZAoptc,lm)
	 . i $G(ZAoptc("lm")) k RAlm(lm)
	 . q
	 m4_setCatIsbdMemberships(RAlm,PDtarget)
	 q

	def IN(PDsource, PDtarget):
	 n x,y,z,RAin
	 m4_getCatIsbdFullTexts(RAin,PDsource)
	 m4_setCatIsbdFullTexts(RAin,PDtarget)
	 q

	def DR(PDsource, PDtarget):
	 n x,y,z
	 n RAdr
	 m4_getCatIsbdCarriers(RAdr,PDsource)
	 m4_setCatIsbdCarriers(RAdr,PDtarget)
	 q
	 ;

	def LG(PDsource, PDtarget):
	 n x,y,z
	 n RAlg
	 m4_getCatIsbdLanguages(RAlg,PDsource)
	 m4_setCatIsbdLanguages(RAlg,PDtarget)
	 q
	 ;

	def TI(PDsource, PDtarget):
	 n x,y,z
	 n RAti
	 m4_getCatIsbdTitles(RAti,PDsource)
	 m4_setCatIsbdTitles(RAti,PDtarget,keywords="copy")
	 q
	 ;

	def AU(PDsource, PDtarget):
	 n x,y,z
	 n RAau
	 m4_getCatIsbdAuthors(RAau,PDsource)
	 m4_setCatIsbdAuthors(RAau,PDtarget)
	 q

	def CA(PDsource, PDtarget):
	 n x,y,z
	 n RAca
	 m4_getCatIsbdCorporateAuthors(RAca,PDsource)
	 m4_setCatIsbdCorporateAuthors(RAca,PDtarget)
	 q

	def ED(PDsource, PDtarget):
	 n x,y,z
	 n RAed
	 m4_getCatIsbdEditions(RAed,PDsource)
	 m4_setCatIsbdEditions(RAed,PDtarget)
	 q

	def IM(PDsource, PDtarget):
	 n x,y,z
	 n RAim
	 m4_getCatIsbdImpressums(RAim,PDsource)
	 m4_setCatIsbdImpressums(RAim,PDtarget)
	 q

	def CO(PDsource, PDtarget):
	 n x,y,z
	 n RAco
	 m4_getCatIsbdCollations(RAco,PDsource)
	 m4_setCatIsbdCollations(RAco,PDtarget)
	 q

	def NT(PDsource, PDtarget):
	 n x,y,z
	 n RAnt
	 m4_getCatIsbdNotes(RAnt,PDsource)
	 m4_setCatIsbdNotes(RAnt,PDtarget)
	 q

	def IF(PDsource, PDtarget):
	 n x,y,z,RAif
	 m4_getCatGenInfo(RAif,PDsource)
	 m4_setCatGenInfo(RAif,PDtarget)
	 q

	def NR(PDsource, PDtarget, PAopt):
	 n x,y,z
	 n RAnr
	 m4_getCatIsbdNumbers(RAnr,PDsource)
	 i '$D(PAopt("nr","*")) d
	 . s x=""
	 . f  s x=$O(RAnr(x)) q:x=""  d
	 .. s y=$g(RAnr(x,"ty"))
	 .. i y="" k RAnr(x) q
	 .. i '$g(PAopt("nr",y)) k RAnr(x) q
	 .. q
	 . q
	 m4_setCatIsbdNumbers(RAnr,PDtarget)
	 q

	def SU(PDsource, PDtarget):
	 n x,y,z
	 n RAow
	 m4_getCatContSubjects(RAow,PDsource)
	 m4_setCatContSubjects(RAow,PDtarget)
	 q

	def RL(PDsource, PDtarget):
	 n x,y,z,RAold,RAnew,rec,sc,ty
	 s RAold("ty")="",RAold("sc")="",RAold("rec")=""
	 m4_nextCatRelation(RAnew,PDtarget,RAold)
	 i $g(RAnew("rec"))="" d
	 . s RAold("ty")="",RAold("sc")="",RAold("rec")=""
	 . s rec=""
	 . f  d  q:rec=""
	 .. k RAnew
	 .. m4_nextCatRelation(RAnew,PDsource,RAold)
	 .. s rec=$G(RAnew("rec")) q:rec=""
	 .. s sc=$G(RAnew("sc"))
	 .. s ty=$G(RAnew("ty"))
	 .. m RAold=RAnew
	 .. q:ty=""
	 .. m4_addCatRelation(PDtarget,ty,sc,rec)
	 .. q
	 q

	def %cpHol(PDploi, PDcloi, PDins):
	 n x,y,z,return,fromins,fromcloi,fromploi,RAhols1,RAhols,acq,toins,tocloi
	 s return=""
	 s fromploi=PDploi
	 s tocloi=PDcloi
	 s toins=PDins
	 m4_getInstellingAndCatRecordFromHolding(fromins,fromcloi,fromploi)
	 ;test ploi bestaat
	 i fromcloi="" q "ploi onbekend^"_fromploi
	 i fromins="" q "ploi(1) onbekend^"_fromploi
	 k RAhols
	 s RAhols(fromploi)=""
	 m4_getCatPkHoldings(RAhols,fromcloi,fromins,0)
	 ;copy kan beginnen
	 s acq=""
	 k RAhols1
	 m RAhols1=RAhols(fromploi)
	 m4_setCatPkHolding(acq,RAhols1,tocloi,toins,"")
	 s return=$g(RAhols1("pkn"))
	 q return

	def cpHolIs(PDcloif, PDcloit, PDins, PAispat):
	 n x,y,z,RAlib,RAholf,ploif,ploin,match
	 s match=0
	 i $d(PAispat) d  q:'match
	 . s match=0,y=""
	 . f  s y=$O(PAispat(y)) q:y=""  s z=PAispat(y) i PDins?@z s match=1 q
	 . q
	 m4_getCatPkHoldings(RAholf,PDcloif,PDins,0)
	 s ploif=""
	 f  s ploif=$O(RAholf(ploif)) q:ploif=""  d
	 . m4_copyCatHolding(ploin,ploif,PDcloit,PDins)
	 . q
	 q
	 ;

	def cpHolAll(PDcloif, PDcloit, PAispat):
	 n x,y,z,RAlib,RAholf,ploif,ploin,match,ins
	 m4_getCatPkLibraries(RAlib,PDcloif)
	 s ins=""
	 f  s ins=$O(RAlib(ins)) q:ins=""  d
	 . d cpHolIs(PDcloif,PDcloit,ins,.PAispat)
	 . q
	 q`

	needle := "m4_getCatIsbdTitles"
	blob := []byte(data)
	tableb, tableg, tablef := BMCreateTable([]byte(needle))
	ok := BMSearch(blob, []byte(needle), tableb, tableg, tablef)

	if !ok {
		t.Errorf("Should have found a match: %v", ok)
		return
	}

	needle = "m4_getCatIsbdTitleS"
	tableb, tableg, tablef = BMCreateTable([]byte(needle))
	ok = BMSearch(blob, []byte(needle), tableb, tableg, tablef)

	if ok {
		t.Errorf("Should not have found a match: %v", ok)
		return
	}

}

func TestAbout19(t *testing.T) {
	data := []byte(`loi tg:
xyx
`)
	expected := data

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
	aboutline := AboutLine(about)
	if aboutline != "" {
		t.Errorf("Error: aboutline : \n`%s`\n\nexpected:\n`Hello`\n", aboutline)
	}
}
func TestAbout01(t *testing.T) {
	data := []byte(`a
b
c
`)

	about := About(data)
	if string(data) != string(about) {
		t.Errorf("Error: about: `%s`", string(about))
	}
}

func TestAbout02(t *testing.T) {
	data := []byte(`"""
b
c
`)

	about := About(data)
	if string(data) != string(about) {
		t.Errorf("Error: about: `%s`", string(about))
	}
}

func TestAbout03(t *testing.T) {
	data := []byte(`

"""
b
c
`)

	about := About(data)
	if string(data) != string(about) {
		t.Errorf("Error: about: `%s`", string(about))
	}
}

func TestAbout04(t *testing.T) {
	data := []byte(`"""
b
c
"""
`)
	expected := []byte(`//
//b
//c
//
`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestAbout05(t *testing.T) {
	data := []byte(`"""
b
c
"""`)
	expected := []byte(`//
//b
//c
//`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestAbout06(t *testing.T) {
	data := []byte(`"""
b
c
"""


A
B`)
	expected := []byte(`//
//b
//c
//


A
B`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestAbout07(t *testing.T) {
	data := []byte(`"""
b
c
"""


A
B
`)
	expected := []byte(`//
//b
//c
//


A
B
`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestAbout08(t *testing.T) {
	data := []byte(`

	"""a b"""

A
https://anet.be
B
`)
	expected := []byte(`

//a b

A
https://anet.be
B
`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestAbout09(t *testing.T) {
	data := []byte(`"""
 About: Hello
"""

 // a b c`)
	expected := []byte(`//
// About: Hello
//

 // a b c`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
	aboutline := AboutLine(about)
	if aboutline != "Hello" {
		t.Errorf("Error: aboutline : \n`%s`\n\nexpected:\n`Hello`\n", aboutline)
	}
}

func TestAbout099(t *testing.T) {
	data := []byte(`// About: Hell
//
 // a b c`)

	about := About(data)

	aboutline := AboutLine(about)
	if aboutline != "Hello" {
		t.Errorf("Error: aboutline : \n`%s`\n\nexpected:\n`Hello`\n", aboutline)
	}
}

func TestAbout10(t *testing.T) {
	data := []byte(`""" about: Hello
"""

 // a b c`)
	expected := []byte(`// about: Hello
//

 // a b c`)

	about := About(data)
	if string(expected) != string(about) {
		t.Errorf("Error: about: \n`%s`\n\nexpected:\n`%s`\n", string(about), string(expected))
	}
}

func TestBlobSplit(t *testing.T) {
	s := `Hello World
A
B
screen xyz
a
b
format $uvw
g
h`

	parts := BlobSplit([]byte(s), []string{"screenx"}, false)
	if len(parts) != 1 {
		t.Errorf("Error: there should be 1 part: %d found\n\n[%s]\n", len(parts), string(parts[0]))
	}

	parts = BlobSplit([]byte(s), []string{"screen"}, false)

	if len(parts) != 2 {
		t.Errorf("Error: there should be 2 parts: %d found\n\n[%s]\n", len(parts), string(parts[0]))
	}
	if !strings.HasPrefix(string(parts[1]), "screen ") {
		t.Errorf("Error: part[1] should start with \n[%s]\n", string(parts[1]))
	}

	parts = BlobSplit([]byte(s), []string{"screen", "format"}, false)

	if len(parts) != 3 {
		t.Errorf("Error: there should be 3 parts: %d found\n\n[%s]\n", len(parts), string(parts[0]))
	}
	if !strings.HasPrefix(string(parts[1]), "screen ") {
		t.Errorf("Error: part[1] should start with \n[%s]\n", string(parts[1]))
	}

	if !strings.HasPrefix(string(parts[2]), "format ") {
		t.Errorf("Error: part[2] should start with \n[%s]\n", string(parts[2]))
	}

}

func TestIgnore(t *testing.T) {
	data := []byte(`abc<ignore> </ignore>ABC`)
	ignore := Ignore(data)
	if string(ignore) != "abcABC" {
		t.Errorf("Error: found: [%s]", string(ignore))
	}

	data = []byte(`abcABC`)
	ignore = Ignore(data)
	if string(ignore) != "abcABC" {
		t.Errorf("Error: found: [%s]", string(ignore))
	}
	data = []byte(`abc<ignore> ABC`)
	ignore = Ignore(data)
	if string(ignore) != "abc" {
		t.Errorf("Error: found: [%s]", string(ignore))
	}
	data = []byte(`abc<ignore> </ignore>ABC<ignore>D</ignore>E`)
	ignore = Ignore(data)
	if string(ignore) != "abcABCE" {
		t.Errorf("Error: found: [%s]", string(ignore))
	}
}

func TestBuidArgs01(t *testing.T) {
	datas := []string{
		`($p(RAdetAd(x,i,"st"),m4_CRLF)_$c(1)_x_":post:gn"_i)`,
		`($g(FDid("layout","set",iset,x)))XYZ`,
		`(check, old="\u005E", new="^", all=1)`,
		`($exist=«exist», $reason=«x», $palnr=«RDpal»)`,
		`(   )`,
		`($palnr=«RDpaln», $paltype=«FDid», $test=«"1"»)`,
		`("","pallet lijst vervallen","["_$s(lsx="lsb":"barcodes",lsx="lso":"object",1:"scan")_"]")`,
		`(x,ZAtmp,del=m4_CRLF,seq=1,uniq=1)`,
	}
	for _, data := range datas {
		xargs, until, msg := BuildArgs(data)
		if true {
			t.Errorf("\n\nError:\ndata:%s\nlen(xargs):%d\nxargs:%s\nuntil:%s\nmsg:%s\n", data, len(xargs), strings.Join(xargs, "|"), until, msg)
		}
	}
}

func TestIsObjStarter(t *testing.T) {
	type T struct {
		S string
		R bool
	}

	TestData := []T{
		{
			"",
			false,
		},
		{
			"Hello World",
			false,
		},
		{
			"m4_",
			false,
		},
		{
			"4_",
			false,
		},
		{
			"m4_A",
			true,
		},
		{
			"i4_A",
			true,
		},
		{
			"l4_A",
			false,
		},
		{
			"l4_N",
			false,
		},
		{
			"l4_Nphp",
			false,
		},
		{
			"l4_N_",
			false,
		},
		{
			"l4_Nphp_",
			false,
		},
		{
			"l4_N_a",
			true,
		},
		{
			"l4_Nphp_a",
			true,
		},
		{
			"l4_Nphp__",
			false,
		},
		{
			"r4_N",
			false,
		},
		{
			"r4_n",
			true,
		},
		{
			"r4__b",
			false,
		},
		{
			"m4_m4",
			true,
		},
		{
			"m4_i4",
			true,
		},
		{
			"m4_i4_a",
			false,
		},
		{
			"m4_i4_a",
			false,
		},
		{
			"m4_l4_a_aaa",
			true,
		},
	}

	for _, test := range TestData {
		blob := []byte(test.S)
		r := IsObjStarter(blob)

		if r != test.R {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound: %v\nexpected: %v", string(blob), r, test.R))
			return
		}
	}

}
func TestSplitter01(t *testing.T) {
	type T struct {
		S string
		R []string
	}

	TestData := []T{
		{
			"",
			[]string{""},
		},
		{
			"q",
			[]string{"q"},
		},
		{
			"Hello World",
			[]string{"Hello World"},
		},
		{
			"4_Hello",
			[]string{"4_Hello"},
		},
		{
			"4_",
			[]string{"4_"},
		},
		{
			"m4_",
			[]string{"m4_"},
		},
		{
			"m4_A",
			[]string{"", "m4_A", ""},
		},
		{
			"m4__",
			[]string{"m4__"},
		},
		{
			"m4_ABC(",
			[]string{"", "m4_ABC", "("},
		},
		{
			"Qm4_ABC(",
			[]string{"Q", "m4_ABC", "("},
		},
		{
			"m4_ABC(",
			[]string{"", "m4_ABC", "("},
		},
		{
			"m4_ABCm4_DEF",
			[]string{"", "m4_ABC", "", "m4_DEF", ""},
		},
		{
			"m4_ABCr4_ab_c_d_m4_DEF",
			[]string{"", "m4_ABC", "", "r4_ab_c_d_", "", "m4_DEF", ""},
		},
		{
			"Hellom4_ABCr4_ab_c_d_m4_DEF World",
			[]string{"Hello", "m4_ABC", "", "r4_ab_c_d_", "", "m4_DEF", " World"},
		},
		{

			"m4_ABCl4_Njs_H1there:World",
			[]string{"", "m4_ABC", "", "l4_Njs_H1there", ":World"},
		},
		{

			"m4_ABCl4_N_H1there:World",
			[]string{"", "m4_ABC", "", "l4_N_H1there", ":World"},
		},
		{

			"m4_m4",
			[]string{"", "m4_m4", ""},
		},
		{

			"m4_m4_m4",
			[]string{"m4_", "m4_m4", ""},
		},
		{

			"m4_m4_m4_m4",
			[]string{"", "m4_m4", "_", "m4_m4", ""},
		},
	}

	for _, test := range TestData {
		blob := []byte(test.S)
		r := ObjectSplitter(blob)
		rs := make([]string, 0)
		for _, part := range r {
			rs = append(rs, string(part))
		}

		rj, _ := json.MarshalIndent(rs, "", "    ")
		Rj, _ := json.MarshalIndent(test.R, "", "    ")

		srj := string(rj)
		sRj := string(Rj)

		if srj != sRj {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound:\n%s\n\nexpected:\n%s", test.S, string(rj), string(Rj)))
			return
		}
	}
}

func TestObjName(t *testing.T) {
	type T struct {
		S  string
		OK bool
	}

	TestData := []T{
		{
			"",
			false,
		},
		{
			"m4_CO",
			true,
		},
		{
			"am4_CO",
			false,
		},
		{
			"m4_CO?",
			false,
		},
		{
			"m4_COm4_AB",
			false,
		},
	}

	for _, x := range TestData {
		ok := IsObjectName(x.S)
		if ok != x.OK {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound:\n%t\n\nexpected:\n%t", x.S, ok, x.OK))
		}
	}

}

func TestJoiner(t *testing.T) {
	type T struct {
		S string
		R string
	}

	TestData := []T{
		{
			"",
			"",
		},
		{
			" ",
			" ",
		},
		{
			"32",
			" ",
		},
		{
			"Hello",
			"Hello",
		},
		{
			"13,10",
			"\r\n",
		},
		{
			"10,A",
			"10,A",
		},
	}

	for _, x := range TestData {
		s := Joiner(x.S)
		if x.R != s {
			t.Errorf(fmt.Sprintf("\n\n%s\n\nfound:\n%s\n\nexpected:\n%s", x.S, s, x.R))
		}
	}

}

func TestDecomment(t *testing.T) {

	x := "Hello https://world // en de rest"
	buf := Decomment([]byte(x))
	y := buf.String()
	if y != "Hello https://world " {
		t.Errorf(y)
	}
	x = `$$keygen: s RAkeyref("id")=$p(RDmtloi,":",3),RAkeyref("url")="https://dev.anet.be/desktop/"_RAkeyref("id")`
	buf = Decomment([]byte(x))
	y = buf.String()
	if y != x {
		t.Errorf(y)
	}
}
