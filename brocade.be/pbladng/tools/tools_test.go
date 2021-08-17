package tools

import (
	"strings"
	"testing"
	"time"
)

func TestIsUTF8(t *testing.T) {

	body := []byte("Hello\n")
	body = append(body, 129)
	body = append(body, []byte("World\n")...)

	err := IsUTF8(body, 1)

	if err == nil {
		t.Errorf("Problem: should have an error")
		return
	}
	if !strings.Contains(err.Error(), "line 2:") {
		t.Errorf("Problem: should contain line 2:")
		return
	}
	body = nil
	err = IsUTF8(body, 1)
	if err != nil {
		t.Errorf("Problem: should not have an error")
		return
	}

	body = []byte("Hello\n")
	body = append(body, []byte("World\n")...)

	err = IsUTF8(body, 1)
	if err != nil {
		t.Errorf("Problem: should not have an error2")
		return
	}
}

func TestPhone(t *testing.T) {
	type teststruct struct {
		test   string
		expect string
	}

	tests := []teststruct{
		{
			test:   "09 3 8 5 6 2 0 3",
			expect: "{(09) 385 62 03}",
		},
		{
			test:   " 04   7 8 28 20   26",
			expect: " {(0478) 28 20 26}",
		},
		{
			test:   " {04   7 8 28 20   26}",
			expect: " {04   7 8 28 20   26}",
		},
		{
			test:   " 04   7 8 28 20   ",
			expect: " 04   7 8 28 20   ",
		},
		{
			test:   "09 3 8 5 6 2 0 3  of 04   7 8 28 20   26",
			expect: "{(09) 385 62 03}",
		},
	}

	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc := Phone(work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestShowPhone(t *testing.T) {
	type teststruct struct {
		test   string
		expect string
	}

	tests := []teststruct{
		{
			test:   "0 9 3 8 5 6 2 0 3",
			expect: "(09) 385 62 03",
		},
		{
			test:   " 0 4   7 8 28 20   26",
			expect: "(0478) 28 20 26",
		},
	}

	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc := showphone(work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestEuro(t *testing.T) {
	type teststruct struct {
		test   string
		expect string
	}

	tests := []teststruct{
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
			expect: "Euro World",
		},
		{
			test:   "Euro",
			expect: "Euro",
		},
		{
			test:   "World 15,3    Euro",
			expect: "World {15,3 EUR}",
		},
		{
			test:   "World 15    Euro",
			expect: "World {15 EUR}",
		},
		{
			test:   "World 15,3    \nEuro",
			expect: "World 15,3    \nEuro",
		},

		{
			test:   "15 Europa",
			expect: "15 Europa",
		},
	}

	for _, test := range tests {
		work := test.test
		expect := test.expect
		calc := Euro(work)
		if expect != calc {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\ncalc:`%s`\n\n", work, expect, calc)
			return

		}
	}
}

func TestNumberSplit(t *testing.T) {
	type teststruct struct {
		test   string
		before string
		number string
		after  string
		money  bool
	}

	tests := []teststruct{
		{
			test:   "",
			before: "",
			number: "",
			after:  "",
		},
		{
			test:   "Hello World",
			before: "Hello World",
			number: "",
			after:  "",
		},
		{
			test:   "15Hello World",
			before: "",
			number: "15",
			after:  "Hello World",
		},
		{
			test:   "Hello15World",
			before: "Hello",
			number: "15",
			after:  "World",
		},
		{
			test:   "Hello World15",
			before: "Hello World",
			number: "15",
			after:  "",
		},
		{
			test:   "{15}Hello World",
			before: "{15}Hello World",
			number: "",
			after:  "",
		},
		{
			test:   "Hello{15}World",
			before: "Hello{15}World",
			number: "",
			after:  "",
		},
		{
			test:   "Hello World{15}",
			before: "Hello World{15}",
			number: "",
			after:  "",
		},
		{
			test:   "Hello World{15}1ABC",
			before: "Hello World{15}",
			number: "1",
			after:  "ABC",
		},
		{
			test:   `Hello World\{15\}1ABC`,
			before: "Hello World\\{",
			number: "15",
			after:  `\}1ABC`,
		},
		{
			test:   "Hello15,3World",
			before: "Hello",
			number: "15,3",
			money:  true,
			after:  "World",
		},
		{
			test:   "Hello15,World",
			before: "Hello",
			number: "15",
			money:  true,
			after:  ",World",
		},
	}

	for _, test := range tests {
		work := test.test
		money := test.money
		before, number, after := NumberSplit(work, money, 0)
		if test.before != before || test.number != number || test.after != after {
			t.Errorf("Problem:\ntest:`%s`\ncalc.before:`%s`\ncalc.number:`%s`\ncalc.after:`%s`\n\n", work, before, number, after)
		}
	}
}

func TestLeftTrim(t *testing.T) {
	type teststruct struct {
		test   string
		number int
		after  string
	}

	tests := []teststruct{
		{
			test:   "",
			number: 0,
			after:  "",
		},
		{
			test:   "Hello World",
			number: 0,
			after:  "Hello World",
		},
		{
			test:   "     Hello World",
			number: 0,
			after:  "Hello World",
		},
		{
			test:   " \n Hello World",
			number: 1,
			after:  "Hello World",
		},
		{
			test:   "\n\n",
			number: 2,
			after:  "",
		},
	}

	for _, test := range tests {
		work := test.test
		after, number := LeftTrim(work)
		if test.after != after || test.number != number {
			t.Errorf("Problem:\ntest:`%s`\ncalc.number:`%d`\ncalc.after:`%s`\n\n", work, number, after)
		}
	}
}

func TestLeftWord(t *testing.T) {
	type teststruct struct {
		test   string
		before string
		after  string
	}

	tests := []teststruct{
		{
			test:   "",
			before: "",
			after:  "",
		},
		{
			test:   "Hello",
			before: "Hello",
			after:  "",
		},
		{
			test:   "Hello World",
			before: "Hello",
			after:  " World",
		},
		{
			test:   "     Hello World",
			before: "",
			after:  "     Hello World",
		},
		{
			test:   "a1 Hello World",
			before: "a",
			after:  "1 Hello World",
		},
		{
			test:   "Hello\n\n",
			before: "Hello",
			after:  "\n\n",
		},
	}

	for _, test := range tests {
		work := test.test
		before, after := LeftWord(work)
		if test.after != after || test.before != before {
			t.Errorf("Problem:\ntest:`%s`\ncalc.before:`%s`\ncalc.after:`%s`\n\n", work, before, after)
		}
	}
}

func TestFirstAlfa(t *testing.T) {
	type teststruct struct {
		test   string
		before string
		word   string
		after  string
	}

	tests := []teststruct{
		{
			test:   "",
			before: "",
			word:   "",
			after:  "",
		},
		{
			test:   "Hello",
			before: "",
			word:   "Hello",
			after:  "",
		},
		{
			test:   "Hello World",
			before: "",
			word:   "Hello",
			after:  " World",
		},
		{
			test:   "     1Hello World",
			before: "     ",
			word:   "1Hello",
			after:  " World",
		},
		{
			test:   "a1 Hello World",
			before: "",
			word:   "a1",
			after:  " Hello World",
		},
		{
			test:   "\n\nHello\n\n",
			before: "\n\n",
			word:   "Hello",
			after:  "\n\n",
		},
	}

	for _, test := range tests {
		work := test.test
		before, word, after := FirstAlfa(work)
		if test.after != after || test.before != before || test.word != word {
			t.Errorf("Problem:\ntest:`%s`\ncalc.before:`%s`\ncalc.after:`%s`\ncalc.word:`%s`\n\n", work, before, after, word)
		}
	}
}

func TestLastAlfa(t *testing.T) {
	type teststruct struct {
		test   string
		before string
		word   string
		after  string
	}

	tests := []teststruct{
		{
			test:   "",
			before: "",
			word:   "",
			after:  "",
		},
		{
			test:   "Hello",
			before: "",
			word:   "Hello",
			after:  "",
		},
		{
			test:   "Hello World",
			before: "Hello ",
			word:   "World",
			after:  "",
		},
		{
			test:   "Hello World1     ",
			before: "Hello ",
			word:   "World1",
			after:  "     ",
		},
		{
			test:   "a1 Hello World b1",
			before: "a1 Hello World ",
			word:   "b1",
			after:  "",
		},
		{
			test:   "\n\nHello\n\n",
			before: "\n\n",
			word:   "Hello",
			after:  "\n\n",
		},
	}

	for _, test := range tests {
		work := test.test
		before, word, after := LastAlfa(work)
		if test.after != after || test.before != before || test.word != word {
			t.Errorf("Problem:\ntest:`%s`\ncalc.before:`%s`\ncalc.after:`%s`\ncalc.word:`%s`\n\n", work, before, after, word)
		}
	}
}

func TestParseIsoDate(t *testing.T) {
	type teststruct struct {
		test  string
		found string
		err   bool
	}

	tests := []teststruct{
		{
			test:  "",
			found: "",
			err:   true,
		},
		{
			test:  "Hello",
			found: "",
			err:   true,
		},
		{
			test:  "2021-15-30",
			found: "",
			err:   true,
		},
		{
			test:  "2021-02-29",
			found: "",
			err:   true,
		},
		{
			test:  "2021-7-7",
			found: "2021-07-07",
			err:   false,
		},
	}

	for _, test := range tests {
		work := test.test
		found, err := ParseIsoDate(work)
		if test.err && err == nil {
			t.Errorf("Problem:\ntest:`%s`\nshould give an error", work)
		}
		if !test.err && test.found != found.Format(time.RFC3339)[:10] {
			t.Errorf("Problem:\ntest:`%s`\nshould give " + found.Format(time.RFC3339)[:10])
		}
	}
}
