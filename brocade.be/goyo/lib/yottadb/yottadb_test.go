package yottadb

import "testing"

func TestMakeGlobalRef(t *testing.T) {
	type test struct {
		ref  string
		erro bool
		glvn string
	}

	tests := []test{

		{
			ref:  "",
			erro: true,
			glvn: "",
		},
		{
			ref:  "^A1",
			erro: false,
			glvn: "^A1",
		},
		{
			ref:  "/A1",
			erro: false,
			glvn: "^A1",
		},
		{
			ref:  "/A1/abc",
			erro: false,
			glvn: `^A1("abc")`,
		},
	}

	for _, mytest := range tests {
		ref := mytest.ref
		glvn := mytest.glvn
		erro := mytest.erro
		fglvn, err := Glvn(ref)

		if erro != (err != nil) || glvn != fglvn {
			t.Errorf("\n%#v\n", mytest)
			t.Errorf("fglvn=%s glvn=%s ref=%s\n", fglvn, glvn, ref)
			if err != nil {
				t.Errorf("%s\n", err.Error())
			}
		}

	}
}
