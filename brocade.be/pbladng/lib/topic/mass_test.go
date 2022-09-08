package topic

import (
	"testing"
)

func TestPlace(t *testing.T) {
	type teststruct struct {
		test   string
		expect string
	}

	tests := []teststruct{
		{
			test:   "eke",
			expect: "eke",
		},
		{
			test:   "Eke",
			expect: "eke",
		},
		{
			test:   "Asper",
			expect: "asper",
		},
		{
			test:   "de pinte",
			expect: "depinte",
		},
		{
			test:   "abc",
			expect: "",
		},
	}
	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc, _ := findPlace(work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return
		}
	}

}

func TestPerson(t *testing.T) {
	type teststruct struct {
		test   string
		expect string
	}

	tests := []teststruct{
		{
			test:   "adg",
			expect: "ADG",
		},
		{
			test:   "ADG",
			expect: "ADG",
		},
		{
			test:   "Annemie De Gussem",
			expect: "ADG",
		},
	}
	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc, _ := findPerson("nazareth", work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return
		}
	}

}
