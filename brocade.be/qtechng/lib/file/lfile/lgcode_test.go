package lfile

import (
	"bytes"
	"strings"
	"testing"

	qmumps "brocade.be/base/mumps"
	//qsource "brocade.be/qtechng/lib/source"
)

func TestLgcode01(t *testing.T) {
	r := "9.99"
	data := []byte(`lgcode lg1:
		N: «Hello»
		E: Hi
`)
	lgcode := &Lgcode{
		ID:      "testcide",
		Source:  "/test/a/b/c.l",
		Version: r,
	}
	lgcode.Loads(data)
	mumps := lgcode.Mumps("batch")
	if len(mumps) != 15 {
		t.Errorf("Error: %s", mumps)
	}

}

func TestLgcode02(t *testing.T) {
	r := "9.99"
	data := []byte(`lgcode lg1:
		N: «Hel
lo»
		E: Hi
`)
	lgcode := &Lgcode{
		ID:      "testcide",
		Source:  "/test/a/b/c.l",
		Version: r,
	}
	lgcode.Loads(data)
	mumps := lgcode.Mumps("batch")
	buf := new(bytes.Buffer)
	qmumps.Println(buf, mumps)
	result := buf.String()
	if !strings.Contains(result, "$C(10)") {
		t.Errorf("Error: %s", result)
	}

}

func TestLgcode03(t *testing.T) {
	r := "9.99"
	data := []byte(`lgcode lg1:
		N: «Heléélo»
		E: Hi
`)
	lgcode := &Lgcode{
		ID:      "testcide",
		Source:  "/test/a/b/c.l",
		Version: r,
	}
	lgcode.Loads(data)
	mumps := lgcode.Mumps("batch")
	buf := new(bytes.Buffer)
	qmumps.Println(buf, mumps)
	result := buf.String()
	if !strings.Contains(result, "_$C(195,169,195,169)_") {
		t.Errorf("Error: %s", result)
	}

}

func TestLgcode04(t *testing.T) {
	r := "9.99"
	data := []byte(`lgcode lg1:
		N: «Helé»
		E: Hi
`)
	lgcode := &Lgcode{
		ID:      "testcide",
		Source:  "/test/a/b/c.l",
		Version: r,
	}
	lgcode.Loads(data)
	mumps := lgcode.Mumps("batch")
	buf := new(bytes.Buffer)
	qmumps.Println(buf, mumps)
	result := buf.String()
	if !strings.Contains(result, "_$C(195,169)") {
		t.Errorf("Error: %s", result)
	}
}

func TestLgcode05(t *testing.T) {
	r := "9.99"
	data := []byte(`lgcode lg1:
		N: «é»
		E: Hi
`)
	lgcode := &Lgcode{
		ID:      "testcide",
		Source:  "/test/a/b/c.l",
		Version: r,
	}
	lgcode.Loads(data)
	mumps := lgcode.Mumps("batch")
	buf := new(bytes.Buffer)
	qmumps.Println(buf, mumps)
	result := buf.String()
	if !strings.Contains(result, "$C(195,169)_") {
		t.Errorf("Error: %s", result)
	}
}
