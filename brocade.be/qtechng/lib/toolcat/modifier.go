package toolcat

import (
	"encoding/json"
	"strconv"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Modifier struct {
	Name        string      `json:"modifier"`
	Description string      `json:"description"`
	Expand      bool        `json:"expand"`
	Number      interface{} `json:"number"`
	WithDefault bool        `json:"withdefault"`
	Default     interface{} `json:"default"`
	Type        string      `json:"type"`
	ES          ES          `json:"extrastring"`
	EM          ExtraM      `json:"extramap"`
}

func (modifier Modifier) ModifierYaml() (m *qyaml.Node) {

	m = &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}

	if modifier.Name == "" {
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
			Value: "Modifiers",
			Tag:   "!!str",
		},
		this)

	name := modifier.Name

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
			Value: strings.TrimSpace(modifier.Description),
			Tag:   "!!str",
		},
	)

	if modifier.Expand {
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

	switch nr := modifier.Number.(type) {
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

	switch df := modifier.Default.(type) {
	case string:
		if strings.TrimSpace(df) != "" {
			modifier.WithDefault = true
		}
	case nil:
	default:
		modifier.WithDefault = true
	}

	if modifier.WithDefault {
		switch df := modifier.Default.(type) {

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
	nature := strings.ToLower(strings.TrimSpace(modifier.Type))
	if nature != "" {
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
	modifier.ES.ESYaml(props)
	modifier.EM.EMYaml(props)

	return
}

func (arg *Modifier) Load(s string) error {
	return json.Unmarshal([]byte(s), arg)
}

func (modifier Modifier) String() string {
	m := modifier.ModifierYaml()
	s := yaml(m)
	lines := strings.SplitN(s, "\n", -1)
	j := -1
	for i, line := range lines {
		line := strings.TrimSpace(line)
		if strings.HasPrefix(line, "Modifiers:") {
			if len(line) == i+1 {
				return ""
			}
			j = i
		}
	}
	if j != -1 {
		s = strings.Join(lines[j+1:], "\n")
	}
	return s
}
