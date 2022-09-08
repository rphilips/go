package tools

import (
	"encoding/json"
	"testing"
)

func TestColon(t *testing.T) {
	type linestruct struct {
		line   string
		expect string
	}

	lines := []linestruct{
		{
			line:   "// a:b",
			expect: "// a:b",
		},
		{
			line:   "a b",
			expect: "a b",
		},
		{
			line:   "a:b:c",
			expect: "a: b: c",
		},
		{
			line:   "a:b https://w3c.org A:B a:1",
			expect: "a: b https://w3c.org A: B a:1",
		},
	}

	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := FixColon(work)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestDelim(t *testing.T) {
	type linestruct struct {
		line   string
		start  int
		sub    string
		expect string
	}

	lines := []linestruct{
		{
			line:   "// a:b",
			start:  3,
			sub:    "a",
			expect: "// a:b",
		},
		{
			line:   "abc",
			start:  1,
			sub:    "b",
			expect: "a*b*c",
		},
		{
			line:   "abc",
			start:  2,
			sub:    "c",
			expect: "ab*c*",
		},
		{
			line:   "abc",
			start:  0,
			sub:    "a",
			expect: "*a*bc",
		},

		{
			line:   "abBc",
			start:  1,
			sub:    "bB",
			expect: "a*bB*c",
		},
		{
			line:   "abcC",
			start:  2,
			sub:    "cC",
			expect: "ab*cC*",
		},
		{
			line:   "aAbc",
			start:  0,
			sub:    "aA",
			expect: "*aA*bc",
		},

		{
			line:   "*abBc*",
			start:  2,
			sub:    "bB",
			expect: "*abBc*",
		},
		{
			line:   "*abcC*",
			start:  3,
			sub:    "cC",
			expect: "*abcC*",
		},
		{
			line:   "*aAbc*",
			start:  1,
			sub:    "aA",
			expect: "*aAbc*",
		},
	}

	for _, line := range lines {
		work := line.line
		sub := line.sub
		index := line.start
		delim := '*'
		expect := line.expect
		calc := FixDelim(work, index, sub, delim)
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestGetDelims(t *testing.T) {

	line := " |Hello| World |Moon|  "

	m := make(map[string]bool)
	GetDelims(line, '|', m)
	if len(m) != 2 || !m["Hello"] || !m["Moon"] {
		t.Errorf("Problem:\nline:`%s`\ngot: %v\n", line, m)
		return
	}

	line = " |Hello\\| World |Moon|  "
	m = make(map[string]bool)
	GetDelims(line, '|', m)
	if len(m) != 1 || !m["Hello\\| World"] {
		t.Errorf("Problem:\nline:`%s`\ngot: %v\n", line, m)
		return
	}

}

func TestIndexRune(t *testing.T) {
	type linestruct struct {
		line   string
		expect int
	}

	lines := []linestruct{
		{
			line:   "",
			expect: -1,
		},
		{
			line:   "abc",
			expect: -1,
		},
		{
			line:   "\\{abc",
			expect: -1,
		},
		{
			line:   "a\\{bc",
			expect: -1,
		},
		{
			line:   "ab\\{c",
			expect: -1,
		},
		{
			line:   "{abc",
			expect: 0,
		},
		{
			line:   "a{bc",
			expect: 1,
		},
		{
			line:   "abc{",
			expect: 3,
		},
		{
			line:   "\\\\{abc",
			expect: 2,
		},
		{
			line:   "a\\\\{bc",
			expect: 3,
		},
		{
			line:   "abc\\\\{",
			expect: 5,
		},
		{
			line:   "\\\\\\{abc",
			expect: -1,
		},
		{
			line:   "a\\\\\\{bc",
			expect: -1,
		},
		{
			line:   `abc\{{`,
			expect: 5,
		},
	}
	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := IndexRune(work, '{')
		if expect != calc {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%d`\ncalc:`%d`\n\n", work, expect, calc)
			return

		}
	}

}

func TestProtected(t *testing.T) {
	type linestruct struct {
		line   string
		expect []string
	}

	lines := []linestruct{
		{
			line:   "",
			expect: []string{""},
		},
		{
			line:   "abc",
			expect: []string{"abc"},
		},
		{
			line:   "\\{abc",
			expect: []string{"\\{abc"},
		},
		{
			line:   "a\\{bc",
			expect: []string{"a\\{bc"},
		},
		{
			line:   "ab\\{c",
			expect: []string{"ab\\{c"},
		},
		{
			line:   "{abc",
			expect: []string{"{abc"},
		},
		{
			line:   "a{bc",
			expect: []string{"a{bc"},
		},
		{
			line:   "abc{",
			expect: []string{"abc{"},
		},
		{
			line:   "{abc}",
			expect: []string{"", "{abc}", ""},
		},
		{
			line:   "A{abc}C",
			expect: []string{"A", "{abc}", "C"},
		},
		{
			line:   "A{abc}{}C",
			expect: []string{"A", "{abc}", "", "{}", "C"},
		},
		{
			line:   "A{abc\\}}{}C",
			expect: []string{"A", "{abc\\}}", "", "{}", "C"},
		},
		{
			line:   "A{abc\\}}{\\}C",
			expect: []string{"A", "{abc\\}}", "{\\}C"},
		},
	}
	for _, line := range lines {
		work := line.line
		expect := line.expect
		calc := Protected(work)
		ej, _ := json.Marshal(expect)
		cj, _ := json.Marshal(calc)
		es := string(ej)
		cs := string(cj)
		if es != cs {
			t.Errorf("Problem:\nline:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, es, cs)
			return

		}
	}

}
