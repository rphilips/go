package strings

import "testing"

func TestLatin1(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "abc",
			expect: "abc",
		},
		{
			line:   "",
			expect: "",
		},
		{
			line:   "één ©",
			expect: "één ©",
		},
		{
			line:   "αβγ",
			expect: "abg",
		},
		{
			line:   "€ 50",
			expect: "EUR 50",
		},
		{
			line:   " € 50",
			expect: " EUR 50",
		},

		{
			line:   " €50 ",
			expect: " EUR 50 ",
		},
		{
			line:   " €50",
			expect: " EUR 50",
		},
		{
			line:   " €50",
			expect: " EUR 50",
		},
		{
			line:   `“”‘’«»“„”`,
			expect: `""''«»",,"`,
		},
		{
			line:   "\u2013\u2014\u2015",
			expect: `-----`,
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := Latin1(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}
