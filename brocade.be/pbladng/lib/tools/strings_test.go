package tools

import (
	"testing"
)

func TestColon2(t *testing.T) {
	s := "**Doctor:**"
	f := Colon(s)
	if s != f {
		t.Errorf("Problem: f: %s", "["+f+"]")
	}

}
