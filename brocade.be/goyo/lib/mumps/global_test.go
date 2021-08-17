package mumps

import (
	"fmt"
	"testing"
)

func TestMakeGlobalRef(t *testing.T) {
	type test struct {
		input  string
		erro   bool
		global bool
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
			input:  "ABC",
			erro:   false,
			global: true,
			gloref: "^ABC",
			subs:   []string{`^ABC`},
		},
		{
			input:  "^%ABC1",
			erro:   false,
			global: false,
			gloref: "%ABC1",
			subs:   []string{`ABC1`},
		},
		{
			input:  "^%ABC1/",
			erro:   false,
			global: false,
			gloref: "%ABC1",
			subs:   []string{`%ABC1`},
		},
		{
			input:  "^%ABC1/q/w",
			erro:   false,
			global: false,
			gloref: `%ABC1("q","w")`,
			subs:   []string{`%ABC1`, `q`, `w`},
		},
		{
			input:  "^%ABC1/\\q/w",
			erro:   false,
			global: false,
			gloref: `%ABC1("\q","w")`,
			subs:   []string{`%ABC1`, `q`, `w`},
		},
		{
			input:  `/^%ABC1/\\q/"w`,
			erro:   false,
			global: false,
			gloref: `ABC1("\q","w")`,
			subs:   []string{`%ABC1`, `q`, `wz`},
		},
	}

	for _, mytest := range tests {
		input := mytest.input
		global := mytest.global
		gloref := mytest.gloref
		// fmt.Println("input:", input)
		subs := mytest.subs
		err := mytest.erro
		g, ss, e := MakeGlobalRef(input, global)

		if g == gloref && cmp(subs, ss) && ((err && e.Error() != "") || (!err && e == nil)) {
			continue
		}

		fmt.Printf("\n%#v\n", mytest)
		fmt.Printf("%s\n", g)
		fmt.Printf("%#v\n", ss)
		fmt.Printf("%#v\n", e)

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
