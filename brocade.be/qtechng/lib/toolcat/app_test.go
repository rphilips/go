package toolcat

import (
	"encoding/json"
	"testing"
	//qsource "brocade.be/qtechng/lib/source"
)

func TestApp(t *testing.T) {
	app := App{
		Title: `I say "Hello World"`,
		Description: `De :mod:calc rekenmachine bevat een aantal functionaliteiten om te
werken met getallen en strings.

Tevens illustreert deze toepassing hoe een *Toolcat applicatie*
kan worden opgezet.`,
		Exe: "calcng",
		C: Condition{
			Os:    []string{"unix", "mac", "linux"},
			Qtech: []string{"W", "B"},
			Registry: map[string]string{
				"os-sep":      "*",
				"qtechng-ssh": "ssh",
				"support-dir": "abcd",
			},
		},
	}
	t.Errorf("YAML: \n\n\n%s", app)
}

func TestJson(t *testing.T) {
	app := App{
		Title: `I say "Hello World"`,
		Description: `De :mod:calc rekenmachine bevat een aantal "functionaliteiten" om te
werken met getallen en strings.

Tevens illustreert deze toepassing hoe een *Toolcat applicatie*
kan worden opgezet:
    - A
	- B
	- C`,
		Exe: "calcng",
		C: Condition{
			Os:    []string{"unix", "mac", "linux"},
			Qtech: []string{"W", "B"},
			Registry: map[string]string{
				"os-sep":      "*",
				"qtechng-ssh": "ssh",
				"support-dir": "abcd",
			},
		},
	}
	json, _ := json.Marshal(app)
	t.Errorf("JSON: \n\n\n%s", string(json))
}

func TestLoadApp(t *testing.T) {

	json := `{"title":"I say \"Hello World\"","exe":"calcng","description":"De :mod:calc rekenmachine bevat een aantal functionaliteiten om te\nwerken met getallen en strings.\n\nTevens illustreert deze toepassing hoe een *Toolcat applicatie*\nkan worden opgezet.","condition":{"os":["unix","mac","linux"],"qtech":["W","B"],"registry":{"os-sep":"*","qtechng-ssh":"ssh","support-dir":"abcd"}}}`
	app := &App{}
	app.Load(json)
	t.Errorf("YAML: \n\n\n%s", app)
}
