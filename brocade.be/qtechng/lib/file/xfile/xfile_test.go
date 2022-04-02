package xfile

import (
	"testing"

	qobject "brocade.be/qtechng/lib/object"
)

func TestParse01(t *testing.T) {
	data := []byte(`// -*- coding: utf-8 -*-
// About: Hello
	
format      a    :
zA
 fB

format      a    :
 zA
  fB
 

screen  c:
A
 B

format      a    :
zA
 fB


text       t  : .....

tA
 tB


format      b    :
 fA
  fB


format      $$b    :
  fA
   fB`)

	xfile := new(XFile)
	xfile.SetRelease("1.11")
	xfile.SetEditFile("hello/world")
	err := qobject.Loads(xfile, data, true)
	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}

	if len(xfile.Widgets) != 7 {
		t.Errorf("Not enough widgets: %d", len(xfile.Widgets))
		return

	}

}
