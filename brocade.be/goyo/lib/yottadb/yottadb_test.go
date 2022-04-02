package yottadb

import "testing"

func TestQS(t *testing.T) {
	type test struct {
		glvn string
		subs []string
	}

	tests := []test{

		{
			glvn: "",
			subs: nil,
		},
		{
			glvn: "^ABC",
			subs: []string{"^ABC"},
		},
		{
			glvn: "^ABC(1)",
			subs: []string{"^ABC", `1`},
		},
		{
			glvn: `^ABC("abc")`,
			subs: []string{"^ABC", `"abc"`},
		},
		{
			glvn: `^ABC($H)`,
			subs: []string{"^ABC", `$H`},
		},
		{
			glvn: `^ABC("ab""c")`,
			subs: []string{"^ABC", `"ab""c"`},
		},
		{
			glvn: `^ABC("ab""c,")`,
			subs: []string{"^ABC", `"ab""c,"`},
		},
		{
			glvn: `^ABC( "ab""c"   `,
			subs: []string{"^ABC", `"ab""c"`},
		},

		{
			glvn: `^ABC( "a b""c"   `,
			subs: []string{"^ABC", `"a b""c"`},
		},
		{
			glvn: `^ABC("abc",XYZ("A", "B"`,
			subs: []string{"^ABC", `"abc"`, `XYZ("A","B")`},
		},
	}

	for _, mytest := range tests {
		ref := mytest.glvn
		esubs := mytest.subs
		csubs := QS(ref)

		if !cmp(esubs, csubs) {
			t.Errorf("\n%s\n", ref)
			t.Errorf("\n%#v\n", esubs)
			t.Errorf("\n%#v\n", csubs)

		}

	}
}

func TestUnQS(t *testing.T) {
	type test struct {
		glvn string
		subs []string
	}

	tests := []test{

		{
			glvn: "",
			subs: nil,
		},
		{
			glvn: "^ABC",
			subs: []string{"^ABC"},
		},
		{
			glvn: "^ABC(1)",
			subs: []string{"^ABC", `1`},
		},
		{
			glvn: `^ABC("abc")`,
			subs: []string{"^ABC", `"abc"`},
		},

		{
			glvn: `^ABC("ab""c")`,
			subs: []string{"^ABC", `"ab""c"`},
		},
		{
			glvn: `^ABC("ab""c,")`,
			subs: []string{"^ABC", `"ab""c,"`},
		},
		{
			glvn: `^ABC(0)`,
			subs: []string{"^ABC", `'$H`},
		},
	}

	for _, mytest := range tests {
		eglvn := mytest.glvn
		subs := mytest.subs
		cglvn := UnQS(subs)

		if eglvn != cglvn {
			t.Errorf("\n%#v\n", subs)
			t.Errorf("\neglvn: %s\n", eglvn)
			t.Errorf("\ncglvn: %s\n", cglvn)

		}

	}
}

func cmp(slice1 []string, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func TestMakeGlobalRef(t *testing.T) {
	type test struct {
		ref  string
		glvn string
	}

	tests := []test{

		{
			ref:  "",
			glvn: "",
		},
		{
			ref:  "^A1",
			glvn: "^A1",
		},
		{
			ref:  "/A1",
			glvn: "^A1",
		},
		{
			ref:  "/A1/abc",
			glvn: `^A1("abc")`,
		},
		{
			ref:  "/A1/abc/",
			glvn: `^A1("abc","")`,
		},
		{
			ref:  "/A1/abc//",
			glvn: `^A1("abc","","")`,
		},
		{
			ref:  "^A1('$H)",
			glvn: "^A(0)",
		},
	}

	for _, mytest := range tests {
		ref := mytest.ref
		glvn := mytest.glvn
		fglvn := Glvn(ref)

		if glvn != fglvn {
			t.Errorf("\n%#v\n", mytest)
			t.Errorf("fglvn=%s glvn=%s ref=%s\n", fglvn, glvn, ref)

		}

	}
}
