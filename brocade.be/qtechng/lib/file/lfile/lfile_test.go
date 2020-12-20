package lfile

import (
	"testing"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
)

func TestParse01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-



	                             // About: Hello


	lgcode warningnumbers:
		N: «Opgelet ! Er staan cijfers in de auteursnaam en dit is GEEN authority code»
	E:          Note! There are numbers in the author name and this is NOT an authority code
		    F         : ⟦«Attention ! Le nom d'auteur contient des chiffres et il ne s'agit
PAS d'une notice d'autorité»⟧`)

	lfile := new(LFile)
	err := qobject.Loads(lfile, data)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	expect := `// -*- coding: utf-8 -*-
//
// About: Hello`
	if expect != lfile.Preamble {
		t.Errorf("Comment found: `%s`", lfile.Preamble)
	}
	lgcodes := lfile.Lgcodes

	if len(lgcodes) != 1 {
		t.Errorf("# lgcodes found: `%d`", len(lgcodes))
	}

	if len(lgcodes) == 1 {
		lgcode := lgcodes[0]

		if lgcode.N != "Opgelet ! Er staan cijfers in de auteursnaam en dit is GEEN authority code" {
			t.Errorf("N-translation: `%s`", lgcode.N)
		}

		if lgcode.E != "Note! There are numbers in the author name and this is NOT an authority code" {
			t.Errorf("E-translation: `%s`", lgcode.E)
		}

		if lgcode.F != `«Attention ! Le nom d'auteur contient des chiffres et il ne s'agit
PAS d'une notice d'autorité»` {
			t.Errorf("F-translation: `%s`", lgcode.F)
		}
	}
}

func TestParse02(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode warningnumbers:
		N: Opgelet

		lgcode warningnumbers2:
		N: Opgelet2

		lgcode warningnumbers:
N: Opgelet2

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)

	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "lfile.lint.double" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lfile.lint.double` error")
	}

}

func TestParse03(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode warningnumbers:
		N: Opgelet`)
	data = append(data, 130)
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "lfile.lint.utf8" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lfile.lint.utf8` error")
	}

}

func TestParse04(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode warningnumbers:
		N: «Opgelet»»

		lgcode warningnumbers2:
		N: Opgelet2

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "lfile.lint.parse" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lfile.lint.parse` error")
	}

}

func TestParse05(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode warningnumbers:

		lgcode warningnumbers2:
		N: Opgelet2

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(qerror.ErrorSlice)[0].(*qerror.QError)
		if e.Ref[0] != "lgcode.lint.empty" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lgcode.lint.empty` error")
	}

}

func TestParse06(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode warningnumbers:
	Alias: xxx.yyy
		N: Opgelet2

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(qerror.ErrorSlice)[0].(*qerror.QError)
		if e.Ref[0] != "lgcode.lint.alias.nonempty" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lgcode.lint.alias.nonempty` error")
	}

}

func TestParse07(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode abc.warningnumbers:
	Alias: 3xxx.yyy`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(qerror.ErrorSlice)[0].(*qerror.QError)
		if e.Ref[0] != "lgcode.lint.alias.form" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lgcode.lint.alias.form` error")
	}

}

func TestParse08(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode abc.warningnumbers:
	Alias: xxx.yyy.scope

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(qerror.ErrorSlice)[0].(*qerror.QError)
		if e.Ref[0] != "lgcode.lint.alias.scope1" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lgcode.lint.alias.scope1` error")
	}

}

func TestParse09(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode ns.warningnumbers.scope:
	Alias: xxx.yyy

`)
	//
	lfile := new(LFile)
	err := qobject.Lint(lfile, data, nil)
	if err != nil {
		e := err.(qerror.ErrorSlice)[0].(*qerror.QError)
		if e.Ref[0] != "lgcode.lint.alias.scope2" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `lgcode.lint.alias.scope2` error")
	}

}

func TestSort01(t *testing.T) {
	data := []byte(`//
	// Hallo

	lgcode ns.warningnumbers.scope:
	   N: ned1

	lgcode ns.warningnumbers:
	   N: ned2

	lgcode abcdef:
	   N: ned3

	lgcode ns.:
	   N: ned3`)

	lfile := new(LFile)
	qobject.Loads(lfile, data)
	lfile.Sort()
	lgcodes := lfile.Lgcodes

	if lgcodes[0].ID != "ns." {
		t.Errorf("First is `%s`", lgcodes[0].ID)
	}

	if lgcodes[1].ID != "ns.warningnumbers" {
		t.Errorf("Second is `%s`", lgcodes[1].ID)
	}
	if lgcodes[2].ID != "ns.warningnumbers.scope" {
		t.Errorf("Third is `%s`", lgcodes[2].ID)
	}

	if lgcodes[3].ID != "abcdef" {
		t.Errorf("Fourth is `%s`", lgcodes[3].ID)
	}

}
