package ifile

import (
	"testing"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
)

func TestParse01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-



	                             // About: Hello


	include hallo1:
		Hallo1
		
	include hallo2:
		« Hallo2 »
		
	include hallo3:
		⟦ Hallo3 ⟧


		include hallo4:
		Hallo4`)

	ifile := new(IFile)
	err := qobject.Loads(ifile, data, true)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	expect := `// -*- coding: utf-8 -*-
//
// About: Hello`
	if expect != ifile.Preamble {
		t.Errorf("Comment found: `%s`", ifile.Preamble)
	}
	includes := ifile.Includes

	if len(includes) != 4 {
		t.Errorf("# includes found: `%d`", len(includes))
	}

	if len(includes) == 4 {
		include := includes[0]

		if include.Content != "Hallo1" {
			t.Errorf("Content: `%s`", include.Content)
		}
		include = includes[1]
		if include.Content != " Hallo2 " {
			t.Errorf("Content: `%s`", include.Content)
		}
		include = includes[2]
		if include.Content != ` Hallo3 ` {
			t.Errorf("Content: `%s`", include.Content)
		}
		include = includes[3]
		if include.Content != `Hallo4` {
			t.Errorf("Content: `%s`", include.Content)
		}
	}
}

func TestParse02(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-



	// About: Hello


include hallo1:
Hallo1

include hallo2:
« Hallo2 »

include hallo2:
⟦ Hallo3 ⟧


include hallo4:
Hallo4`)
	//
	ifile := new(IFile)
	err := qobject.Lint(ifile, data, nil)
	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "ifile.lint.double" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `ifile.lint.double` error")
	}

}

func TestParse03(t *testing.T) {
	data := []byte(`//
	// Hallo

	include hallo1:
Hallo1`)
	data = append(data, 130)
	ifile := new(IFile)
	err := qobject.Lint(ifile, data, nil)

	//
	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "ifile.lint.utf8" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `ifile.lint.utf8` error")
	}

}

func TestParse04(t *testing.T) {
	data := []byte(`//
	// Hallo

	include 1hello:
		Opgelet
`)
	//
	ifile := new(IFile)
	err := qobject.Lint(ifile, data, nil)
	if err != nil {

		e := err.(*qerror.QError)
		if e.Ref[0] != "ifile.lint.parse" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `ifile.lint.parse` error")
	}

}

func TestParse05(t *testing.T) {
	data := []byte(`//
	// Hallo

	include hello1:

`)
	//
	ifile := new(IFile)
	err := qobject.Loads(ifile, data, true)
	if err != nil {
		e := err.(*qerror.QError)
		if e.Ref[0] != "ifile.loads" {
			t.Errorf("Error: %s", err)
		}
	} else {
		t.Errorf("Should have `ifile.loads` error")
	}

}
