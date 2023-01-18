package tools

import "testing"

func TestHour(t *testing.T) {
	type linestruct struct {
		line    string
		expect1 string
		expect2 string
		expect3 string
	}

	lines := []linestruct{
		{
			line:    "11u",
			expect1: "11.00 u.",
			expect2: "11.00 u. 11.00 u.",
			expect3: "a 11.00 u. b",
		},
		{
			line:    " 11u ",
			expect1: " 11.00 u. ",
			expect2: " 11.00 u.   11.00 u. ",
			expect3: "a  11.00 u.  b",
		},
		{
			line:    "11 u",
			expect1: "11.00 u.",
			expect2: "11.00 u. 11.00 u.",
			expect3: "a 11.00 u. b",
		},
		{
			line:    "11.30 u.",
			expect1: "11.30 u.",
			expect2: "11.30 u. 11.30 u.",
			expect3: "a 11.30 u. b",
		},
		{
			line:    "11 u.",
			expect1: "11.00 u.",
			expect2: "11.00 u. 11.00 u.",
			expect3: "a 11.00 u. b",
		},
		{
			line:    "11u30",
			expect1: "11.30 u.",
			expect2: "11.30 u. 11.30 u.",
			expect3: "a 11.30 u. b",
		},
	}

	for _, line := range lines {
		work1 := line.line
		work2 := work1 + " " + work1
		work3 := "a " + work1 + " b"
		expect1 := line.expect1
		expect2 := line.expect2
		expect3 := line.expect3
		calc1 := Hour(work1)
		calc2 := Hour(work2)
		calc3 := Hour(work3)
		if expect1 != calc1 {
			t.Errorf("xProblem:\nline:`%s`\nexpect:`%s`\n calc1:`%s`\n\n", work1, expect1, calc1)
			return
		}
		if expect2 != calc2 {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\n calc2:`%s`\n\n", work2, expect2, calc2)
			return
		}
		if expect3 != calc3 {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\n calc3:`%s`\n\n", work3, expect3, calc3)
			return
		}
	}
}
