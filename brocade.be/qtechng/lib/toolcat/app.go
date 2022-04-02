package toolcat

import (
	"encoding/json"

	qyaml "gopkg.in/yaml.v3"
)

type App struct {
	Title       string    `json:"title"`
	Exe         string    `json:"exe"`
	Description string    `json:"description"`
	C           Condition `json:"condition"`
}

func (app *App) Load(s string) error {
	return json.Unmarshal([]byte(s), app)
}

func (app App) String() string {
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
			Value:       app.Title,
			Tag:         "!!str",
			FootComment: "\n",
		})

	if app.Exe != "" {

		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "Naam voor de executable",
				Tag:   "!!str",
			},
			&qyaml.Node{
				Kind:        qyaml.ScalarNode,
				Value:       app.Exe,
				Tag:         "!!str",
				FootComment: "\n",
			})

	}
	m.Content = append(m.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: "Beschrijving",
			Tag:   "!!str",
		},
		&qyaml.Node{
			Kind:        qyaml.ScalarNode,
			Value:       app.Description,
			Tag:         "!!str",
			FootComment: "\n",
		})

	yc := app.C.CYaml("")
	if len(yc.Content) != 0 {
		m.Content = append(m.Content,
			&qyaml.Node{
				Kind:  qyaml.ScalarNode,
				Value: "Voorwaarden",
				Tag:   "!!str",
			},
			yc)
	}

	s := yaml(m)
	return s
}
