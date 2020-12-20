package mfile

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFormat1(t *testing.T) {
	text := []byte(`"""about: Hello
"""
// a b c



`)
	about := `//about: Hello

// a b c
`
	output := new(bytes.Buffer)
	err := Format("fname", text, output)
	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if about != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[%s]\nExpected: \n[%s]\n", ftext, about))
	}

}

func TestFormat2(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("def    HHH(PDx, PDy):      // Hello")
	expect := "// About: \n\ndef HHH(PDx, PDy):      // Hello\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}

}

func TestFormat3(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("def    HHH:      // Hello")
	expect := "// About: \n\ndef HHH:      // Hello\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}

func TestFormat5(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("0        s x=3     // ")
	expect := "// About: \n\n0        s x=3     //\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}

}

func TestFormat6(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte(".    .   . .d sub")
	expect := "// About: \n\n .... d sub\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}

func TestFormat7(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("            d sub")
	expect := "// About: \n\n d sub\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}

func TestFormat8(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte(" d sub\n\n\n s x=3")
	expect := "// About: \n\n d sub\n\n s x=3\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}

func TestFormat9(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("i4_abc")
	expect := "// About: \n\ni4_abc\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}

func TestFormat10(t *testing.T) {
	output := new(bytes.Buffer)
	text := []byte("m4_abc")
	expect := "// About: \n\nm4_abc\n"
	err := Format("fname", text, output)

	if err != nil {
		t.Errorf("Error:\n%s\n", err)
	}
	ftext := output.String()
	if expect != ftext {
		t.Errorf(fmt.Sprintf("\nFound: \n[[%s]]\nExpected: \n[[%s]]", ftext, expect))
	}
}
