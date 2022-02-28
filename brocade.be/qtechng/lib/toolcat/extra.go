package toolcat

import (
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Extra struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type E []Extra

func (e E) EYaml(m *qyaml.Node) {

	if len(e) == 0 {
		return
	}

	for _, extra := range e {
		key := strings.ToLower(strings.TrimSpace(extra.Key))
		if key == "" {
			continue
		}
		value := strings.TrimSpace(extra.Key)
		if value == "" {
			continue
		}
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: key,
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: value,
				Tag:   "!!str",
			})
	}
}
