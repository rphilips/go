package toolcat

import (
	"encoding/json"
	"sort"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Verb struct {
	Name          string      `json:"name"`
	Title         string      `json:"title"`
	Description   string      `json:"description"`
	C             Condition   `json:"condition"`
	Triggers      []string    `json:"triggers"`
	Examples      []string    `json:"examples"`
	WithArguments bool        `json:"witharguments"`
	WithModifiers bool        `json:"withmodifiers"`
	WithDebug     bool        `json:"withdebug"`
	WithVerbose   bool        `json:"withverbose"`
	A             APIargument `json:"apiargument"`
}

func (verb *Verb) Load(s string) error {
	return json.Unmarshal([]byte(s), verb)
}

func (verb Verb) Signature() string {

	sign := "@toolcat.toolcat\ndef " + verb.Name + "("
	star := true
	count := 0

	if verb.WithArguments {
		sign += "args"
		count++
	}
	if verb.WithModifiers {
		if star {
			star = false
			if count != 0 {
				sign += ", "
			}
			sign += "*"
			count++
		}
		if count != 0 {
			sign += ", "
		}
		sign += "modifiers=None"
		count++
	}
	if verb.WithVerbose {
		if star {
			star = false
			if count != 0 {
				sign += ", "
			}
			sign += "*"
			count++
		}
		if count != 0 {
			sign += ", "
		}
		sign += "verbose=None"
		count++
	}
	if verb.WithDebug {
		if star {
			star = false
			if count != 0 {
				sign += ", "
			}
			sign += "*"
			count++
		}
		if count != 0 {
			sign += ", "
		}
		sign += "debug=None"
		count++
	}
	_, api := verb.A.AYaml()
	if api != "" {
		if star {
			star = false
			if count != 0 {
				sign += ", "
			}
			sign += "*"
			count++
		}
		if count != 0 {
			sign += ", "
		}
		sign += api
		count++
	}
	return sign + ")"
}

func (verb Verb) String() string {
	m := qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}
	m.Content = append(m.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: "Titel",
			Tag:   "!!str",
		},
		&qyaml.Node{
			Kind:        qyaml.ScalarNode,
			Value:       verb.Title,
			Tag:         "!!str",
			FootComment: "\n",
		})

	m.Content = append(m.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: "Beschrijving",
			Tag:   "!!str",
		},
		&qyaml.Node{
			Kind:        qyaml.ScalarNode,
			Value:       verb.Description,
			Tag:         "!!str",
			FootComment: "\n",
		})

	yc := verb.C.CYaml("\n")
	if len(yc.Content) != 0 {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       "Voorwaarden",
				Tag:         "!!str",
				FootComment: "\n",
			},
			yc)
	}
	triggers := verb.Triggers
	if len(triggers) != 0 {
		trig := make([]string, 0)
		for _, t := range triggers {
			t = strings.TrimSpace(t)
			t = strings.ToLower(t)
			if t != "" {
				trig = append(trig, t)
			}
		}
		sort.Strings(trig)
		if len(trig) != 0 {
			m.Content = append(m.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: "Triggers",
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:        qyaml.ScalarNode,
					Value:       strings.Join(trig, ", "),
					Tag:         "!!str",
					FootComment: "\n",
				})
		}

	}
	examples := verb.Examples
	if len(examples) != 0 {
		exmps := make([]string, 0)
		for _, t := range examples {
			t = strings.TrimSpace(t)
			t = strings.ToLower(t)
			if t != "" {
				exmps = append(exmps, t)
			}
		}
		if len(exmps) != 0 {
			ex := &qyaml.Node{
				Kind:    qyaml.SequenceNode,
				Content: []*qyaml.Node{},
			}
			for _, e := range exmps {
				ex.Content = append(ex.Content, &qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: e,
					Tag:   "!!str",
				})
			}

			m.Content = append(m.Content,
				&qyaml.Node{
					Kind:        qyaml.ScalarNode,
					Value:       "Voorbeelden",
					Tag:         "!!str",
					FootComment: "\n",
				},
				ex)
		}

	}

	if !verb.WithArguments {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "Argumenten",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       "Geen argumenten",
				Tag:         "!!str",
				FootComment: "\n",
			})
	}
	if verb.WithArguments {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "Argumenten",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       "",
				Tag:         "!!str",
				FootComment: "\n",
			})
	}

	if verb.WithModifiers {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "Modifiers",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       "",
				Tag:         "!!str",
				FootComment: "\n",
			})
	}
	ya, _ := verb.A.AYaml()
	if len(ya.Content) != 0 {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       "API argumenten",
				Tag:         "!!str",
				FootComment: "\n",
			},
			ya)
	}

	s := yaml(m)
	return s
}
