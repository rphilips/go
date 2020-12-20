package mumps

import (
	"strings"
	"testing"
)

func TestMS01(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{"^ZBCAT", "lvd", "1", "ti"}, "De Witte")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT("lvd","1","ti")="De Witte"` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMS02(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{"^ZBCAT"}, "De Witte")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT="De Witte"` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMS03(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{"^ZBCAT", "A\nB"}, "De Wit\nte")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT("A"_$C(10)_"B")="De Wit"_$C(10)_"te"` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMS04(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{"^ZBCAT", "A\nB"}, "")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT("A"_$C(10)_"B")=""` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMS05(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{"^ZBCAT", "A\nB", "C"}, "«Hello World»")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT("A"_$C(10)_"B","C")=$C(194,171)_"Hello World"_$C(194,187)` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMS06(t *testing.T) {
	mumps := []M{}
	mumps = Set(mumps, []string{`^ZBCAT("a","b")`, "A", "B"}, "«Hello World»")
	b := new(strings.Builder)
	Println(b, mumps)
	x := b.String()
	expect := `s ^ZBCAT("a","b","A","B")=$C(194,171)_"Hello World"_$C(194,187)` + "\n"
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMK01(t *testing.T) {
	mumps := []M{}
	mumps = Kill(mumps, []string{`^ZBCAT("a","b")`})
	b := new(strings.Builder)
	Print(b, mumps)
	x := b.String()
	expect := `k ^ZBCAT("a","b")`
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMK02(t *testing.T) {
	mumps := []M{}
	mumps = Kill(mumps, []string{`^ZBCAT`})
	b := new(strings.Builder)
	Print(b, mumps)
	x := b.String()
	expect := `k ^ZBCAT`
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMK03(t *testing.T) {
	mumps := []M{}
	mumps = Kill(mumps, []string{`^ZBCAT`, "a", "b"})
	b := new(strings.Builder)
	Print(b, mumps)
	x := b.String()
	expect := `k ^ZBCAT("a","b")`
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestMK04(t *testing.T) {
	mumps := []M{}
	mumps = Kill(mumps, []string{`^ZBCAT`, "a\nA", "b"})
	b := new(strings.Builder)
	Print(b, mumps)
	x := b.String()
	expect := `k ^ZBCAT("a"_$C(10)_"A","b")`
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}

func TestME01(t *testing.T) {
	mumps := []M{}
	mumps = Exec(mumps, `s a="A"`)
	b := new(strings.Builder)
	Print(b, mumps)
	x := b.String()
	expect := `s a="A"`
	if x != expect {
		t.Errorf("output: %s %d", x, len(x))
		t.Errorf("expect: %s %d", expect, len(expect))
		return
	}
}
