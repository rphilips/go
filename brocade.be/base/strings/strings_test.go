package strings

import (
	"testing"
)

func TestLeftTrim(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "één ©   ",
			expect: "één ©   ",
		},
		{
			line:   "abc",
			expect: "abc",
		},
		{
			line:   "   αβγ   ",
			expect: "αβγ   ",
		},
		{
			line:   "   abc   ",
			expect: "abc   ",
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := LeftTrimSpace(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestRightTrim(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "één ©   ",
			expect: "één ©",
		},
		{
			line:   "   αβγ   ",
			expect: "   αβγ",
		},
		{
			line:   "abc",
			expect: "abc",
		},
		{
			line:   "   abc   ",
			expect: "   abc",
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := RightTrimSpace(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestLeftRuneString(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "één ©   ",
			expect: "é",
		},

		{
			line:   "abc",
			expect: "a",
		},
		{
			line:   "",
			expect: "",
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := LeftRuneString(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestRightRuneString(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "één ©",
			expect: "©",
		},

		{
			line:   "abc",
			expect: "c",
		},
		{
			line:   "",
			expect: "",
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := RightRuneString(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestReplacer(t *testing.T) {

	keys := map[string]string{"a": "alfa", "b": "beta"}

	work1 := " {a} + {b} + {c} "
	work2 := " {a} + {b} + {c} "
	expect1 := " alfa + beta + {c} "
	expect2 := " alfa + beta + {c} "
	calc1 := Template(work1, keys, "{", "}", "id1")
	if expect1 != calc1 {
		t.Errorf("Problem: id1\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work1, expect1, calc1)
		return

	}
	calc2 := Template(work2, keys, "{", "}", "id2")
	if expect2 != calc2 {
		t.Errorf("Problem: id2\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work2, expect2, calc2)
		return

	}

}
