package tools

import (
	"testing"
	"time"
)

func TestDetect(t *testing.T) {
	now := time.Now()
	times := []string{
		"Maandag *03 oktober 2022* om 13.30 u. CREA-atelier in CC De Brouwerij.",
		"3/10 13.30:A-atelier in CC De Brouwerij.",
	}

	expect := time.Date(2022, 10, 3, 0, 0, 0, 0, now.Location())
	for _, s := range times {
		d := DetectDate(s)
		if d == nil {
			t.Errorf("Problem: d should not be nil")
			return
		}
		if !d.Equal(expect) {
			t.Errorf("Problem: found `%s`", d)
			return
		}
	}
	times = []string{
		"!!! Voor meer info: tel. 09/395.53.71",
	}
	for _, s := range times {
		d := DetectDate(s)
		if d != nil {
			t.Errorf("Problem: d should be nil")
			return
		}

	}
}

func TestString(t *testing.T) {
	type linestruct struct {
		test    string
		expect1 string
		expect2 string
		expect3 string
		after   string
	}

	lines := []linestruct{
		{
			test:    "1-9-2020 Hello",
			expect1: "2020-09-01",
			expect2: "1 september 2020",
			expect3: "dinsdag 1 september 2020",
			after:   " Hello",
		},
		{
			test:    "1/9/2020 Hello",
			expect1: "2020-09-01",
			expect2: "1 september 2020",
			expect3: "dinsdag 1 september 2020",
			after:   " Hello",
		},
		{
			test:    "1 september 2020 Hello",
			expect1: "2020-09-01",
			expect2: "1 september 2020",
			expect3: "dinsdag 1 september 2020",
			after:   " Hello",
		},
		{
			test:    "1 october 2020 Hello",
			expect1: "2020-10-01",
			expect2: "1 oktober 2020",
			expect3: "donderdag 1 oktober 2020",
			after:   " Hello",
		},
		{
			test:    "1 oktober 2020 Hello",
			expect1: "2020-10-01",
			expect2: "1 oktober 2020",
			expect3: "donderdag 1 oktober 2020",
			after:   " Hello",
		},
		{
			test:    "1 oktober Hello",
			expect1: "2022-10-01",
			expect2: "1 oktober 2022",
			expect3: "zaterdag 1 oktober 2022",
			after:   " Hello",
		},
		{
			test:    "29 feb 2020",
			expect1: "2020-02-29",
			expect2: "29 februari 2020",
			expect3: "zaterdag 29 februari 2020",
			after:   "",
		},
		{
			test:    "29 feb '20",
			expect1: "2020-02-29",
			expect2: "29 februari 2020",
			expect3: "zaterdag 29 februari 2020",
			after:   "",
		},
		{
			test:    "'20 feb 29",
			expect1: "2020-02-29",
			expect2: "29 februari 2020",
			expect3: "zaterdag 29 februari 2020",
			after:   "",
		},
		{
			test:    "maandag '20 feb 29",
			expect1: "2020-02-29",
			expect2: "29 februari 2020",
			expect3: "zaterdag 29 februari 2020",
			after:   "",
		},
		{
			test:    "29/6",
			expect1: "2022-06-29",
			expect2: "29 juni 2022",
			expect3: "woensdag 29 juni 2022",
			after:   "",
		},
	}

	for _, line := range lines {
		work := line.test
		expect1 := line.expect1
		expect2 := line.expect2
		expect3 := line.expect3
		after := line.after
		tim, a, err := NewDate(work)
		if err != nil {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nerror:`%s`\n\n", work, expect1, err)
			return
		}
		calc1 := StringDate(tim, "I")
		calc2 := StringDate(tim, "M")
		calc3 := StringDate(tim, "D")
		if calc1 != expect1 {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nfound:`%s`\n", work, expect1, calc1)
			return
		}
		if calc2 != expect2 {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nfound:`%s`\n", work, expect2, calc2)
			return
		}
		if calc3 != expect3 {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nfound:`%s`\n", work, expect3, calc3)
			return
		}
		if a != after {
			t.Errorf("Problem:\ntest:`%s`\nafter:`%s`\nfound:`%s`\n", work, after, a)
		}
	}
}

func TestNew(t *testing.T) {
	type linestruct struct {
		test   string
		err    string
		expect string
	}

	lines := []linestruct{
		{
			test:   "001-00009-2000Hello World",
			err:    "`001-00009-2000Hello World` is not a valid date [-00009-001]",
			expect: "2000-09-01",
		},
		{
			test:   "001-00009-2000 Hello World",
			err:    "",
			expect: "2000-09-01",
		},
		{
			test:   "20 feb '18 llo World",
			err:    "",
			expect: "2018-02-20",
		},
		{
			test:   "29 feb '19 llo World",
			err:    "`29 feb '19 llo World` has not a valid day(29) for month 2",
			expect: "2019-02-29",
		},
		{
			test:   "20 fep '18 llo World",
			err:    "`20 fep '18 llo World` is not a valid date",
			expect: "2018-02-20",
		},
		{
			test:   "20 14 '18 llo World",
			err:    "`20 14 '18 llo World` has not a valid month(14)",
			expect: "2018-02-20",
		},
		{
			test:   "32 maar '18 llo World",
			err:    "`32 maar '18 llo World` has not a valid day(32)",
			expect: "2018-03-32",
		},
		{
			test:   "31 apr '18 llo World",
			err:    "`31 apr '18 llo World` has not 31 days in month 4",
			expect: "",
		},
	}

	for _, line := range lines {
		work := line.test
		expect := line.expect
		e := line.err
		tim, _, err := NewDate(work)
		if err != nil && e != err.Error() {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nerror:`%s`\n\n", work, expect, err)
			return
		}
		if err == nil && e != "" {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\n expected error:`%s`\n\n", work, expect, e)
			return
		}

		if err == nil && expect != StringDate(tim, "I") {
			t.Errorf("Problem:\ntest:`%s`\nexpect:`%s`\nfound:`%s`\n", work, expect, StringDate(tim, "I"))
			return
		}

	}
}
