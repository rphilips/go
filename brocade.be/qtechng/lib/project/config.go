package project

import (
	"encoding/json"
	"sync"

	qerror "brocade.be/qtechng/lib/error"
)

var configCache = new(sync.Map)

// Config structure voor configuratie van een project
type Config struct {
	Passive            bool                `json:"passive"`
	Mumps              []string            `json:"mumps"`
	Groups             []string            `json:"groups"`
	Names              []string            `json:"names"`
	Roles              []string            `json:"roles"`
	VersionLower       string              `json:"versionlower"`
	VersionUpper       string              `json:"versionupper"`
	Py3                bool                `json:"py3"`
	Core               bool                `json:"core"`
	Priority           int                 `json:"priority"`
	NotBrocade         []string            `json:"notbrocade"`
	NotConfig          []string            `json:"notconfig"`
	Binary             []string            `json:"binary"`
	ObjectsNotReplaced map[string][]string `json:"objectsnotreplaced"`
	ObjectsNotChecked  []string            `json:"objectsnotchecked"`
	EmptyDirs          []string            `json:"emptydirs"`
	NotUnique          []string            `json:"notunique"`
}

// IsValid checks if a blob is a validd configuaration

func IsValidConfig(blob []byte) bool {
	cfg := Config{}
	e := json.Unmarshal(blob, &cfg)
	if e != nil {
		return false
	}
	good := map[string]bool{
		"$id":                true,
		"$schema":            true,
		"binary":             true,
		"core":               true,
		"groups":             true,
		"mumps":              true,
		"names":              true,
		"notbrocade":         true,
		"notconfig":          true,
		"objectsnotreplaced": true,
		"passive":            true,
		"priority":           true,
		"py3":                true,
		"roles":              true,
		"versionlower":       true,
		"versionupper":       true,
	}
	m := make(map[string]interface{})
	json.Unmarshal(blob, &m)
	for key := range m {
		if !good[key] {
			return false
		}
	}
	return true
}

// LoadConfig laadt een configuratiebestand
func (project Project) LoadConfig() (config Config, err error) {
	p := project.String()
	r := project.Release().String()
	readonly := project.ReadOnly()
	pid := r + " " + p
	if readonly {
		pid = r + " " + p + " R"
	}
	cfg, _ := configCache.Load(pid)
	if cfg != nil {
		return cfg.(Config), nil
	}
	blob, e := project.Fetch("brocade.json")
	if e != nil {
		return config, e
	}
	e = json.Unmarshal(blob, &config)
	if e != nil {
		err = &qerror.QError{
			Ref:     []string{"config.load.unmarshal"},
			Version: r,
			Project: p,
			Msg:     []string{"Error on unmarshaling configuration file: `" + e.Error() + "'"},
		}
		return config, err
	}
	c, _ := configCache.LoadOrStore(pid, config)
	if c != nil {
		return c.(Config), nil
	}
	err = &qerror.QError{
		Ref:     []string{"config.load.cache"},
		Version: r,
		Msg:     []string{"Cannot load configuration"},
	}
	return
}

// UpdateConfig bewaart een configuratiebestand
func (project Project) UpdateConfig(config Config) {
	p := project.String()
	r := project.Release().String()
	pid := r + " " + p
	configCache.Delete(pid)
	pid = r + " " + p + " R"
	configCache.Delete(pid)
	pid = r + " " + p
	if project.ReadOnly() {
		pid = r + " " + p + " R"
	}
	configCache.LoadOrStore(pid, config)
	return
}
