package toolcat

import (
	"encoding/json"
	"testing"
)

func TestArg(t *testing.T) {
	arg := Arg{
		Name:        "arg*",
		Description: `het eerste argument`,
		Modifier:    `argmodifiers`,
		Number:      `1..10`,
		WithDefault: false,
		Type:        "any",
	}
	t.Errorf("YAML: \n\n\n%s", arg)
}

func TestJsonArg(t *testing.T) {
	arg := Arg{
		Name:        "arg*",
		Description: `het eerste argument`,
		Modifier:    `argmodifiers`,
		Number:      `1..10`,
		WithDefault: false,
		Type:        "any",
	}

	json, _ := json.Marshal(arg)
	t.Errorf("JSON: \n\n\n%s", string(json))
}
