package cmd

import (
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {

	testvalue, err := goversion()
	if err != nil {
		t.Errorf("Running go version fails 1:\n%s", err)
		return
	}
	if !strings.HasPrefix(testvalue, "go1.") {
		t.Errorf("Running go version fails 2:\n%s", testvalue)
		return
	}
}

func TestHostname(t *testing.T) {

	testvalue, err := hostname()
	if err != nil {
		t.Errorf("Getting hostname fails 1:\n%s", err)
		return
	}
	if !strings.HasPrefix(testvalue, "rphilips") {
		t.Errorf("Getting hostname fails 2:\n%s", testvalue)
		return
	}
}
