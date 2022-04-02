package toolcat

import (
	"strconv"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type ExtraS struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

type ExtraM struct {
	Key    string   `json:"key"`
	Values []ExtraS `json:"values"`
}

type ES []ExtraS

func (es ES) ESYaml(m *qyaml.Node) {

	if len(es) == 0 {
		return
	}

	for _, extra := range es {
		key := strings.ToLower(strings.TrimSpace(extra.Key))
		if key == "" {
			continue
		}
		value := extra.Value
		switch v := value.(type) {
		case string:
			value := strings.TrimSpace(v)
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
					Value: v,
					Tag:   "!!str",
				})

		case float64:
			m.Content = append(m.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: key,
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: strconv.Itoa(int(v)),
				})
		}

	}
}

func (em ExtraM) EMYaml(m *qyaml.Node) {
	values := em.Values
	if len(values) == 0 {
		return
	}
	key := strings.TrimSpace(em.Key)
	if key == "" {
		return
	}
	this := &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}
	m.Content = append(m.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		},
		this)

	for _, extra := range values {
		key := strings.ToLower(strings.TrimSpace(extra.Key))
		if key == "" {
			continue
		}
		value := strings.TrimSpace(extra.Value.(string))
		if value == "" {
			continue
		}
		this.Content = append(this.Content,
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
