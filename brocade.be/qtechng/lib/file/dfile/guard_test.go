package dfile

import (
	"encoding/json"
	"fmt"
	"testing"

	qutil "brocade.be/qtechng/lib/util"
)

func TestTokenize03(t *testing.T) {
	s := []byte("true  and        false")
	exp := []string{"true", "false", "and"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}

}

func TestTokenize04(t *testing.T) {
	s := []byte("true  and        (    false  or true ) and true")
	exp := []string{"true", "false", "true", "or", "and"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}

}

func TestTokenize05(t *testing.T) {
	s := []byte("true  or        (    false  and true)  and true")
	exp := []string{"true", "false", "true", "and", "true", "and", "or"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}
}

func TestTokenize06(t *testing.T) {
	s := []byte(`%project sortsAfter "/a/bc"`)
	exp := []string{"%project", `/a/bc`, "sortsAfter"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}
}

func TestTokenize07(t *testing.T) {
	s := []byte(`%project not-sortsAfter "/a/bc"`)
	exp := []string{"%project", `/a/bc`, "not-sortsAfter"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}
}

func TestTokenize08(t *testing.T) {
	s := []byte(`%project not-sortsAfter "/a/bc" and $alfa isInstanceOf "numlit"`)
	exp := []string{"%project", `/a/bc`, "not-sortsAfter", "$alfa", "numlit", "isInstanceOf", "and"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}
}

func TestTokenize09(t *testing.T) {
	s := []byte(`%project not-sortsAfter "/a/bc" and ( $alfa isInstanceOf "numlit" or  $beta    regexpMatches """"   )`)
	exp := []string{"%project", "/a/bc", "not-sortsAfter", "$alfa", "numlit", "isInstanceOf", "$beta", "\"", "regexpMatches", "or", "and"}
	fnd, _ := LoadsGuard(s)

	if !qutil.CmpStringSlice(exp, fnd, false) {
		blob, _ := json.MarshalIndent(fnd, "", "    ")
		t.Errorf(fmt.Sprintf("fnd is `%v`", string(blob)))
	}
}

func TestGuard01(t *testing.T) {
	env := make(map[string]string)

	env["%project"] = "/catalografie/application"
	env["%path"] = "/catalografie/application/export/catexpo.m"
	env["%relpath"] = "export/catexpo.m"
	env["%basename"] = "catexpo.m"
	env["%ext"] = "m"
	env["%version"] = "4.40"

	env["%clib"] = "1"
	env["%mumps"] = "gtm"
	env["%sysname"] = "legato"
	env["%sysgroup"] = "anet"
	env["%osname"] = "linux"

	env["$a"] = `aaa`
	env["$b"] = `RAxyz`
	env["$c"] = `yes`

	type T struct {
		guard string
		exp   bool
	}
	test := []T{
		{`$b contains "Ax"`, true},
		{`$b startsWith "RA"`, true},
		{`$b startsWith "zRA"`, false},
		{`$b endsWith "xyz"`, true},
		{`$b endsWith "xyqz"`, false},
		{`$b endsWith "xyqz"`, false},
		{`$a sortsAfter "bbbb"`, false},
		{`$a sortsAfter "a"`, true},
		{`$a sortsBefore "b"`, true},
		{`$a sortsBefore "a"`, false},
		{`$a isEqualTo "aaa"`, true},
		{`$a isEqualTo "aaaa"`, false},
		{`$a fileMatches "[a-z][a-z][a-z]"`, true},
		{`$a fileMatches "[a-z][a-z][b-z]"`, false},
		{`$a regexpMatches "^[a-z]+"`, true},
		{`$a regexpMatches "^[b-z]+"`, false},
		{`$b isIn "RA"`, false},
		{`$b isIn "aRAxyzb"`, true},
		{`$c isEqualTrueAs "1"`, true},
		{`$c isEqualTrueAs "no"`, false},
	}

	for _, tst := range test {
		postfix, _ := LoadsGuard([]byte(tst.guard))
		calc := Eval(postfix, env)
		if calc != tst.exp {
			t.Errorf(fmt.Sprintf("guard: `%s`\n    expected is `%v`\n    found is `%v`", tst.guard, tst.exp, calc))
		}
	}
}

func TestGuard02(t *testing.T) {
	env := make(map[string]string)

	env["%project"] = "/catalografie/application"
	env["%path"] = "/catalografie/application/export/catexpo.m"
	env["%relpath"] = "export/catexpo.m"
	env["%basename"] = "catexpo.m"
	env["%ext"] = "m"
	env["%version"] = "4.40"

	env["%clib"] = "1"
	env["%mumps"] = "gtm"
	env["%sysname"] = "legato"
	env["%sysgroup"] = "anet"
	env["%osname"] = "linux"

	env["$a"] = `"aaa"`
	env["$b"] = `RAxyz`
	env["$c"] = ``
	env["$d"] = `-123`
	env["$e"] = `12E3`
	env["$f"] = `"abc`
	env["$g"] = `%1`
	env["$h"] = `1a`
	env["$i"] = `RAname("1")`
	env["$j"] = `^RAname("1")`
	env["$k"] = `+^RAname("1")`
	env["$l"] = `@^RAname("1")`
	env["$m"] = `.@abc`

	type T struct {
		guard string
		exp   bool
	}

	test := []T{
		{`$c isInstanceOf "empty"`, true},
		{`$a isInstanceOf "empty"`, false},
		{`$d isInstanceOf "intlit"`, true},
		{`$a isInstanceOf "intlit"`, false},
		{`$e isInstanceOf "numlit"`, true},
		{`$a isInstanceOf "numlit"`, false},
		{`$a isInstanceOf "strlit"`, true},
		{`$w isInstanceOf "strlit"`, false},
		{`$f isInstanceOf "strlit"`, false},
		{`$g isInstanceOf "name"`, true},
		{`$h isInstanceOf "name"`, false},
		{`$i isInstanceOf "lvn"`, true},
		{`$j isInstanceOf "lvn"`, false},
		{`$i isInstanceOf "gvn"`, false},
		{`$j isInstanceOf "gvn"`, true},
		{`$i isInstanceOf "glvn"`, true},
		{`$j isInstanceOf "glvn"`, true},
		{`$i isInstanceOf "expritem"`, false},
		{`$k isInstanceOf "expritem"`, true},
		{`$g isInstanceOf "actualname"`, true},
		{`$l isInstanceOf "actualname"`, true},
		{`$b isInstanceOf "actual"`, true},
		{`$m isInstanceOf "actual"`, true},
	}

	for _, tst := range test {
		postfix, _ := LoadsGuard([]byte(tst.guard))
		calc := Eval(postfix, env)
		if calc != tst.exp {
			t.Errorf(fmt.Sprintf("guard: `%s`\n    expected is `%v`\n    found is `%v`", tst.guard, tst.exp, calc))
		}
	}

}
