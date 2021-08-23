package mumps

import (
	"testing"
)

func TestMakeGlobalRef(t *testing.T) {
	type test struct {
		input  string
		erro   bool
		gloref string
		subs   []string
	}

	tests := []test{

		{
			input:  "",
			erro:   true,
			gloref: "",
			subs:   nil,
		},
		{
			input:  "---",
			erro:   true,
			gloref: "",
			subs:   nil,
		},
		{
			input:  "/ABC",
			erro:   false,
			gloref: "^ABC",
			subs:   []string{`^ABC`},
		},
		{
			input:  "^%ABC1",
			erro:   false,
			gloref: "^%ABC1",
			subs:   []string{`^%ABC1`},
		},
		{
			input:  "^%ABC1/",
			erro:   false,
			gloref: `^%ABC1("")`,
			subs:   []string{`^%ABC1`, ""},
		},
		{
			input:  "^%ABC1/q/w",
			erro:   false,
			gloref: `^%ABC1("q","w")`,
			subs:   []string{`^%ABC1`, `q`, `w`},
		},
		{
			input:  "^%ABC1/\\q/w",
			erro:   false,
			gloref: `^%ABC1("\q","w")`,
			subs:   []string{`^%ABC1`, `\q`, `w`},
		},
		{
			input:  `/%ABC1/\\q/"w`,
			erro:   false,
			gloref: `^%ABC1("\q","""w")`,
			subs:   []string{`^%ABC1`, `\q`, `"w`},
		},
		{
			input:  `/%ABC1/\\q/123`,
			erro:   false,
			gloref: `^%ABC1("\q",123)`,
			subs:   []string{`^%ABC1`, `\q`, `123`},
		},

		{
			input:  `/%ABC1/\\q/10E2`,
			erro:   false,
			gloref: `^%ABC1("\q",1000)`,
			subs:   []string{`^%ABC1`, `\q`, `1000`},
		},
	}

	for _, mytest := range tests {
		input := mytest.input
		gloref := mytest.gloref
		// fmt.Println("input:", input)
		subs := mytest.subs
		err := mytest.erro
		g, ss, e := GlobalRef(input)

		if g == gloref && cmp(subs, ss) && ((err && e.Error() != "") || (!err && e == nil)) {
			continue
		}
		t.Errorf("\n%#v\n", mytest)
		t.Errorf("%s\n", g)
		t.Errorf("%#v\n", ss)
		t.Errorf("%#v\n", e)

	}
}

func cmp(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == len(b) && len(a) == 0 {
		return true
	}
	for i, x := range a {
		if x != b[i] {
			return false
		}
	}
	return true
}
