package tools

import (
	"testing"
)

func TestEuro(t *testing.T) {
	type testeuro struct {
		test   string
		expect string
	}

	tests := []testeuro{
		{
			test:   "",
			expect: "",
		},
		{
			test:   "Hello World",
			expect: "Hello World",
		},
		{
			test:   "HelloEuro World",
			expect: "HelloEuro World",
		},
		{
			test:   "Euro World",
			expect: " EUR World",
		},
		{
			test:   "Euro",
			expect: " EUR ",
		},
		{
			test:   "{Euro}",
			expect: "{Euro}",
		},
		{
			test:   "\\{Euro\\}",
			expect: "\\{ EUR \\}",
		},
		{
			test:   "WorldEuro",
			expect: "WorldEuro",
		},
		{
			test:   "World Euro",
			expect: "World EUR ",
		},
		{
			test:   "World 19Euro",
			expect: "World 19 EUR ",
		},
		{
			test:   "Eur Euro",
			expect: " EUR EUR ",
		},
	}

	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc := euro(work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestNumberSplit(t *testing.T) {
	type testeuro struct {
		test   string
		before string
		number int
		after  string
	}

	tests := []testeuro{
		{
			test:   "",
			before: "",
			number: -1,
			after:  "",
		},
		{
			test:   "Hello World",
			before: "Hello World",
			number: -1,
			after:  "",
		},
		{
			test:   "15Hello World",
			before: "",
			number: 15,
			after:  "Hello World",
		},
		{
			test:   "Hello15World",
			before: "Hello",
			number: 15,
			after:  "World",
		},
		{
			test:   "Hello World15",
			before: "Hello World",
			number: 15,
			after:  "",
		},
	}

	for _, test := range tests {
		work := test.test
		before, number, after := NumberSplit(work)
		if test.before != before || test.number != number || test.after != after {
			t.Errorf("Problem:\ntest:`%s`\ncalc.before:`%s`\ncalc.number:`%d`\ncalc.after:`%s`\n\n", work, before, number, after)
		}
	}
}
