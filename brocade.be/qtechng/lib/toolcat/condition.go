package toolcat

import (
	"sort"
	"strings"

	qyaml "gopkg.in/yaml.v3"
)

type Condition struct {
	Os       []string          `json:"os"`
	Qtech    []string          `json:"qtech"`
	UID      []string          `json:"uid"`
	GID      []string          `json:"gid"`
	M        []string          `json:"m"`
	Registry map[string]string `json:"registry"`
	Sysname  []string          `json:"sysname"`
	Sysgroup []string          `json:"sysgroup"`
	Time     []string          `json:"time"`
	Cwd      []string          `json:"cwd"`
}

func (cond Condition) CYaml(foot string) *qyaml.Node {
	m := &qyaml.Node{
		Kind:    qyaml.MappingNode,
		Content: []*qyaml.Node{},
	}
	worker(cond.Os, "os", strings.ToLower, m)
	worker(cond.Qtech, "qtech", strings.ToUpper, m)
	worker(cond.UID, "uid", nil, m)
	worker(cond.GID, "gid", nil, m)
	worker(cond.M, "m", strings.ToLower, m)
	worker(cond.Sysname, "sysname", nil, m)
	worker(cond.Sysgroup, "sysgroup", nil, m)
	worker(cond.Time, "time", nil, m)
	worker(cond.Cwd, "cwd", nil, m)

	if len(cond.Registry) != 0 {

		regs := make([]string, 0)
		for key, value := range cond.Registry {
			if value == "" {
				continue
			}
			regs = append(regs, key)
		}
		if len(regs) == 0 {
			return m
		}

		r := &qyaml.Node{
			Kind:    qyaml.MappingNode,
			Content: []*qyaml.Node{},
		}
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "registry",
				Tag:   "!!str",
			},
			r)

		sort.Strings(regs)

		for _, reg := range regs {
			r.Content = append(r.Content,
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: reg,
					Tag:   "!!str",
				},
				&qyaml.Node{
					Kind:  qyaml.ScalarNode,
					Value: cond.Registry[reg],
					Tag:   "!!str",
				})
		}
	}
	if foot != "" && len(m.Content) != 0 {
		l := len(m.Content)
		m.Content[l-1].FootComment = foot
	}
	return m
}
