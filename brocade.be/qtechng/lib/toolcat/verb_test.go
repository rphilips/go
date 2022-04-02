package toolcat

import (
	"encoding/json"
	"testing"
	//qsource "brocade.be/qtechng/lib/source"
)

func TestVerb(t *testing.T) {
	verb := Verb{
		Name:        "time",
		Title:       `Bereken de tijd`,
		Description: `Bereken de ISO tijd van nu`,
		C: Condition{
			Os:    []string{"unix", "mac", "linux"},
			Qtech: []string{"W", "B"},
			Registry: map[string]string{
				"os-sep":      "*",
				"qtechng-ssh": "ssh",
				"support-dir": "abcd",
			},
		},
		Triggers:      []string{"tim", "now"},
		Examples:      []string{"calcng time", "calcng now"},
		WithArguments: true,
		WithModifiers: true,
		WithDebug:     true,
		WithVerbose:   true,
		A: []Param{
			{
				Arg:         "zone",
				Description: "timezode"},
			{
				Arg:         "daylight",
				Description: "yes"},
		},
	}
	t.Errorf("YAML: \n\n\n%s", verb)
}

func TestJsonVerb(t *testing.T) {
	verb := Verb{
		Name:        "time",
		Title:       `Bereken de tijd`,
		Description: `Bereken de ISO tijd van nu`,
		C: Condition{
			Os:    []string{"unix", "mac", "linux"},
			Qtech: []string{"W", "B"},
			Registry: map[string]string{
				"os-sep":      "*",
				"qtechng-ssh": "ssh",
				"support-dir": "abcd",
			},
		},
		Triggers:      []string{"tim", "now"},
		Examples:      []string{"calcng time", "calcng now"},
		WithArguments: true,
		WithModifiers: true,
		WithDebug:     true,
		WithVerbose:   true,
		A: []Param{
			{
				Arg:         "zone",
				Description: "timezode"},
			{
				Arg:         "daylight",
				Description: "yes"},
		},
	}

	json, _ := json.Marshal(verb)
	t.Errorf("JSON: \n\n\n%s", string(json))
}
