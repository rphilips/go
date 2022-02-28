package toolcat

import (
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
