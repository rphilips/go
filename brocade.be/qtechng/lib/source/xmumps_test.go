package source

import (
	"testing"
)

func TestXdecomment(t *testing.T) {
	tests := [][3]string{
		{
			"abcd",
			"abcd",
			"",
		},
		{
			"// hello",
			"",
			"// hello",
		},
		{
			"://hello",
			"://hello",
			"",
		},
		{
			"://hello //abcd",
			"://hello ",
			"//abcd",
		},
		{
			"",
			"",
			"",
		},
		{
			`a://hello "//abcd"`,
			`a://hello "//abcd"`,
			"",
		},
		{
			`a://hello "//abcd"//XYZ`,
			`a://hello "//abcd"`,
			"//XYZ",
		},
	}

	for _, test := range tests {
		x, y := xdecomment([]byte(test[0]))
		sx := string(x)
		sy := string(y)

		if test[1] != sx && test[2] != sy {
			t.Errorf("\nTested: [%s]\n    line   : [%s]\n    comment: [%s]\n", test[0], sx, sy)
		}

	}

}
