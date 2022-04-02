package project

import (
	"testing"
)

func TestConfig01(t *testing.T) {
	blob := []byte("{}")
	valid := IsValidConfig(blob)

	if !valid {
		t.Errorf("Is Valid")
		return
	}

	blob = []byte("{          }")
	valid = IsValidConfig(blob)

	if !valid {
		t.Errorf("Is Valid")
		return
	}

	blob = []byte("{          ")
	valid = IsValidConfig(blob)

	if valid {
		t.Errorf("Is Not Valid")
		return
	}

	blob = []byte("")
	valid = IsValidConfig(blob)

	if valid {
		t.Errorf("Is Not Valid")
		return
	}

	blob = []byte(`{"core": true}`)
	valid = IsValidConfig(blob)

	if !valid {
		t.Errorf("Is Valid")
		return
	}

	blob = []byte(`{"core": 1}`)
	valid = IsValidConfig(blob)

	if valid {
		t.Errorf("Is Not Valid")
		return
	}

	blob = []byte(`{"mumps": ["gtm"]}`)
	valid = IsValidConfig(blob)

	if !valid {
		t.Errorf("Is Valid")
		return
	}

	blob = []byte(`{"mumps": "gtm"}`)
	valid = IsValidConfig(blob)

	if valid {
		t.Errorf("Is Not Valid")
		return
	}

	blob = []byte(`{"xmumps": "gtm"}`)
	valid = IsValidConfig(blob)

	if valid {
		t.Errorf("Is Not Valid")
		return
	}

}
