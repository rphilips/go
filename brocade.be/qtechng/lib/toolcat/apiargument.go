package toolcat

import (
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Param struct {
	Arg         string `json:"arg"`
	Description string `json:"description"`
}

type APIargument []Param

func (apiargument APIargument) AYaml() (m *qyaml.Node, sign string) {

	m = &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}

	if len(apiargument) == 0 {
		return m, ""
	}

	for _, param := range apiargument {
		arg := strings.TrimSpace(param.Arg)
		if arg == "" {
			continue
		}
		if sign != "" {
			sign += ", "
		}
		sign += arg + "=None"
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: arg,
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: strings.TrimSpace(param.Description),
				Tag:   "!!str",
			})
	}
	return
}
