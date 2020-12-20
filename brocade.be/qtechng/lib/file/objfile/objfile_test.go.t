package objfile

import (
	"fmt"
	"testing"
)

func TestParsem(t *testing.T) {
	file := "cat.d"
	_, err := Parse(file)
	if err != nil {
		t.Errorf(fmt.Sprintf("Error is:\n %s", err))
	}

}

func TestParsel(t *testing.T) {
	file := "catalografie.l"
	_, err := Parse(file)
	if err != nil {
		t.Errorf(fmt.Sprintf("Error is:\n %s", err))
	}

}

func TestParsei(t *testing.T) {
	file := "bctech.i"
	_, err := Parse(file)
	if err != nil {
		t.Errorf(fmt.Sprintf("Error is:\n %s", err))
	}

}
