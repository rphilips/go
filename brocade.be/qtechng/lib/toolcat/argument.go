package toolcat

import (
	"encoding/json"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Arg struct {
	Name        string `json:"arg"`
	Description string `json:"description"`
	Modifier    string `json: "modifier"`
	Expand      bool   `json: "expand"`
	Number      string `json: "number"`
	WithDefault bool   `json: "withdefault"`
	Default     string `json: "default"`
	Type        string `json: "type"`
	E           E      `json: "extra`
}

func (arg Arg) ArgYaml() (m *qyaml.Node) {

	m = &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}

	if arg.Name == "" {
		return nil
	}
	this := &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}
	props := &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}
	m.Content = append(m.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: "Argumenten",
			Tag:   "!!str",
		},
		this)

	name := arg.Name
	if strings.ContainsRune(name, '*') {
		name = "arg*"
	} else {
		name = strings.TrimLeft(name, " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
		name = "arg" + name
	}

	this.Content = append(this.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: name,
			Tag:   "!!str",
		},
		props)

	props.Content = append(props.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: "betekenis",
			Tag:   "!!str",
		},
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: strings.TrimSpace(arg.Description),
			Tag:   "!!str",
		},
	)

	mod := strings.TrimSpace(arg.Modifier)

	if mod != "" {
		props.Content = append(props.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "modifier",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: mod,
				Tag:   "!!str",
			},
		)
	}

	if arg.Expand {
		props.Content = append(props.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "expandeer",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "ja",
				Tag:   "!!str",
			},
		)
	}

	deflt := strings.TrimSpace(arg.Default)

	if strings.TrimSpace(deflt) != "" {
		arg.WithDefault = true
	}

	if arg.WithDefault {
		props.Content = append(props.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "default",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: arg.Default,
				Tag:   "!!str",
			},
		)
	}
	nature := strings.ToLower(strings.TrimSpace(arg.Type))
	if nature != "" && nature != "string" {
		props.Content = append(props.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "type",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: nature,
				Tag:   "!!str",
			},
		)
	}
	arg.E.EYaml(props)

	return
}

func (arg *Arg) Load(s string) error {
	return json.Unmarshal([]byte(s), arg)
}

func (arg Arg) String() string {
	m := arg.ArgYaml()
	s := yaml(m)
	lines := strings.SplitN(s, "\n", -1)
	for i, line := range lines {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, "Argumenten:") {
			if len(line) == i+1 {
				return ""
			}
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return s
}
