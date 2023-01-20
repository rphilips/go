package tools

import "testing"

func TestBank(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "Be 56 1234567890 12",
			expect: "BE56 1234 5678 9012",
		},
	}

	for _, line := range lines {
		work1 := line.line
		expect1 := line.expect
		calc1 := Bank(work1, false)
		if expect1 != calc1 {
			t.Errorf("xProblem:\nline:`%s`\nexpect:`%s`\n calc1:`%s`\n\n", work1, expect1, calc1)
			return
		}
	}
}
