package source

import (
	"strings"
	"testing"
)

func TestMtransform(t *testing.T) {
	tests := [][3]string{
		{
			"",
			"",
			"",
		},
		{
			"",
			"; hello",
			" ; hello",
		},
		{
			"ABC    ",
			"",
			"ABC ;",
		},
		{
			"ABC    ",
			"; Hello",
			"ABC ; Hello",
		},
		{
			"def ABC :   ",
			"",
			"ABC ;def",
		},
		{
			"def ABC (a, xx,   z) :   ",
			"",
			"ABC(a,xx,z) ;def",
		},
		{
			"    . . s x=1    s z=1   ",
			";hello world",
			" .. s x=1 s z=1 ;hello world",
		},
		{
			".    . . s x=1    q",
			";hello world",
			" ... s x=1 q  ;hello world",
		},
		{
			".    . . s x=1    q",
			"",
			" ... s x=1 q",
		},
		{
			"           s x=1    q",
			"",
			" s x=1 q",
		},
		{
			"     s x=1    s z=1   ",
			";hello world",
			" s x=1 s z=1 ;hello world",
		},
		{
			"     s x=1    q",
			";hello world",
			" s x=1 q  ;hello world",
		},
	}

	for _, test := range tests {
		line := []byte(test[0])
		comment := []byte(test[1])

		mcalc := string(mtransform(line, comment))

		if test[2] != mcalc {
			t.Errorf("\n\nLine: [%s]\nComm: [%s]\nFound: [%s]\n\n", line, comment, mcalc)
		}
	}

}

func TestMbeautify(t *testing.T) {
	tests := [][3]string{
		{
			"",
			"",
			"false",
		},
		{
			"s x=1",
			"s x=1",
			"true",
		},
		{
			`s x=" "`,
			`s x=" "`,
			"true",
		},
		{
			`s x=" ""  "`,
			`s x=" ""  "`,
			"true",
		},
		{
			"s x=1 s y=1",
			"s x=1 s y=1",
			"true",
		},
		{
			"s x=1      s y=1",
			"s x=1 s y=1",
			"true",
		},
		{
			"s x=1      s y=1        s z=1   ",
			"s x=1 s y=1 s z=1",
			"true",
		},
		{
			"d",
			"d",
			"false",
		},
		{
			"d ",
			"d",
			"false",
		},
		{
			"d    q",
			"d  q",
			"false",
		},
		{
			"d    q",
			"d  q",
			"false",
		},
		{
			`d:" " x    q:" "`,
			`d:" " x q:" "`,
			"false",
		},
	}

	for _, test := range tests {
		sx := test[0]
		sy, ok := mbeautify(sx)

		if test[1] != sy {
			t.Errorf("\nTested: [%s]\nFound : [%s]\n", sx, sy)
		}
		if (ok && test[2] == "false") || (!ok && test[2] == "true") {
			t.Errorf("\nTested: [%s]\nFound : [%s]\nok:%t", sx, sy, ok)
		}

	}

}

func TestMdetag(t *testing.T) {
	tests := [][4]string{
		{
			"",
			"",
			"",
			"",
		},
		{
			"          ",
			"",
			"",
			"",
		},
		{
			"ABC",
			"ABC",
			"",
			"",
		},
		{
			"ABC   s x=1",
			"ABC",
			"s x=1",
			"",
		},
		{
			"ABC(a,b , c)   s x=1",
			"ABC(a,b,c)",
			"s x=1",
			"",
		},
		{
			"ABC (a,b , c)   s x=1",
			"ABC(a,b,c)",
			"s x=1",
			"",
		},
		{
			"ABC (    )   s x=1",
			"ABC()",
			"s x=1",
			"",
		},
		{
			"ABC()   s x=1",
			"ABC()",
			"s x=1",
			"",
		},
		{
			"    .  ",
			"",
			".",
			"",
		},
		{
			"      ",
			"",
			"",
			"",
		},
	}

	for _, test := range tests {
		x, y, sz := mdetag([]byte(test[0]))
		sx := string(x)
		sy := string(y)

		if test[1] != sx || test[2] != sy || sz != test[3] {
			t.Errorf("\nTested: [%s]\n    name: [%s]\n    rest: [%s]\n    tag : [%s]\n", test[0], sx, sy, sz)
		}

		x, y, sz = mdetag([]byte("def  " + test[0]))
		sx = string(x)
		sy = string(y)

		if test[1] != sx || test[2] != sy || sz != "def" {
			t.Errorf("\nTested: [%s]\n    name: [%s]\n    rest: [%s]\n    tag : [%s]\n", "def  "+test[0], sx, sy, sz)
		}

	}

}

func TestMdecomment(t *testing.T) {
	tests := [][3]string{
		{
			"abcd",
			"abcd",
			"",
		},
		{
			";",
			"",
			";",
		},
		{
			"abcd/efgh",
			"abcd/efgh",
			"",
		},
		{
			"/efgh",
			"/efgh",
			"",
		},
		{
			"abcd;efgh",
			"abcd",
			";efgh",
		},
		{
			"abcd;",
			"abcd",
			";",
		},
		{
			"abcd/",
			"abcd/",
			"",
		},
		{
			`abcd"i;j"kl`,
			`abcd"i;j"kl`,
			"",
		},
		{
			`abcd"i;j";kl`,
			`abcd"i;j"`,
			";kl",
		},
		{
			`abcd«i;j»;kl`,
			`abcd«i;j»`,
			";kl",
		},
		{
			`abcd«i;j»kl`,
			`abcd«i;j»kl`,
			"",
		},
		{
			`abcd«i;j»klabcd«i;j»;kl`,
			`abcd«i;j»klabcd«i;j»`,
			";kl",
		},
		{
			``,
			``,
			"",
		},
	}

	for _, test := range tests {
		x, y := mdecomment([]byte(test[0]))
		sx := string(x)
		sy := string(y)

		if test[1] != sx && test[2] != sy {
			t.Errorf("\nTested: [%s]\n    line   : [%s]\n    comment: [%s]\n", test[0], sx, sy)
		}

		if strings.Contains(test[0], ";") {
			test[0] = strings.ReplaceAll(test[0], ";", "//")
			test[1] = strings.ReplaceAll(test[1], ";", "//")
			test[2] = strings.ReplaceAll(test[2], ";", "//")
			x, y := mdecomment([]byte(test[0]))
			sx := string(x)
			sy := string(y)

			if test[1] != sx && test[2] != sy {
				t.Errorf("\nTested: [%s]\n    line   : [%s]\n    comment: [%s]\n", test[0], sx, sy)
			}
		}

	}

}
