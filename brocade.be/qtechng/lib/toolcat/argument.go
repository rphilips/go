package toolcat

import (
	"encoding/json"
	"strconv"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Arg struct {
	Name        string      `json:"arg"`
	Description string      `json:"description"`
	Modifier    string      `json:"modifier"`
	Expand      bool        `json:"expand"`
	Number      interface{} `json:"number"`
	WithDefault bool        `json:"withdefault"`
	Default     interface{} `json:"default"`
	Type        string      `json:"type"`
	ES          ES          `json:"extrastring"`
	EM          ExtraM      `json:"extramap"`
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

	if arg.Name == "arg*" {
		switch nr := arg.Number.(type) {
		case string:
			number := strings.TrimSpace(nr)
			if number == "" {
				number = "0.."
			}
			props.Content = append(props.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: "aantal",
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: number,
					Tag:   "!!str",
				},
			)
		case float64:
			props.Content = append(props.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: "aantal",
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: strconv.Itoa(int(nr)),
				},
			)
		}
	}

	switch df := arg.Default.(type) {
	case string:
		if strings.TrimSpace(df) != "" {
			arg.WithDefault = true
		}
	case nil:
	default:
		arg.WithDefault = true
	}

	if arg.WithDefault {
		switch df := arg.Default.(type) {

		case string:
			df = strings.TrimSpace(df)
			props.Content = append(props.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: "default",
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: df,
					Tag:   "!!str",
				},
			)
		case float64:
			props.Content = append(props.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: "default",
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: strconv.Itoa(int(df)),
				},
			)
		}
	}
	nature := strings.ToLower(strings.TrimSpace(arg.Type))
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
	arg.ES.ESYaml(props)
	arg.EM.EMYaml(props)

	return
}

func (arg *Arg) Load(s string) error {
	return json.Unmarshal([]byte(s), arg)
}

func (arg Arg) String() string {
	m := arg.ArgYaml()
	s := yaml(m)
	lines := strings.SplitN(s, "\n", -1)
	j := -1
	for i, line := range lines {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, "Argumenten:") {
			if len(lines) == i+1 {
				return ""
			}
			if j == -1 && line != "" {
				j = i
			}
		}
	}

	if j != -1 {
		s = strings.Join(lines[j+1:], "\n")
	}
	return s
}
