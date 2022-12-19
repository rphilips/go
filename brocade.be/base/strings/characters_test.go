package strings

import (
	"testing"
)

func TestReplacement(t *testing.T) {

	tests := [][2]string{
		{`Hello {a*} en {*b}`, `Hello α en β`},
		{`Hello {a1} en {1b}`, `Hello {a1} en {1b}`},
	}

	for _, test := range tests {

		repl := InsertDiacritic(test[0])

		if repl != test[1] {
			t.Errorf("%s", repl)
		}

	}
}
