package report

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

type header struct {
	Host   string   `json:"host" yaml:"host"`
	Time   string   `json:"time" yaml:"time"`
	Args   []string `json:"args" yaml:"args"`
	Jerror string   `json:"jsonpatherror,omitempty" yaml:"jsonpatherror,omitempty"`
}

type report struct {
	Header header      `json:"ABOUT" yaml:"ABOUT"`
	Result interface{} `json:"DATA" yaml:"DATA"`
	Errors []error     `json:"ERRORS" yaml:"ERRORS"`
}

func Report(r interface{}, e interface{}, jsonpath []string, yaml bool, unquote bool, silent bool) string {
	if silent {
		return ""
	}
	show := report{}

	// header

	host, _ := os.Hostname()
	h := header{
		Host: host,
		Time: time.Now().Format(time.RFC3339),
		Args: os.Args,
	}
	if len(jsonpath) != 0 {
		for _, jp := range jsonpath {
			if jp != "" {
				err := qutil.ParsePath(jp)
				if err != nil {
					h.Jerror = err.Error()

					jsonpath = nil
					break
				}
			}
		}
	}

	show.Header = h

	// result

	show.Result = r

	// Errors
	show.Errors = qerror.FlattenErrors(e)

	// to JSON

	b, _ := json.MarshalIndent(show, "", "    ")
	if show.Errors != nil {
		return string(b)
	}
	s := string(b)

	// Apply jsonpath
	if len(jsonpath) != 0 {
		for _, jp := range jsonpath {
			if jp == "" {
				continue
			}
			s, _ = qutil.JSONpath([]byte(s), jp)
		}
	}
	s = strings.TrimSpace(s)

	if !yaml {
		var x interface{}
		err := json.Unmarshal([]byte(s), &x)
		if err == nil {
			b, _ := json.MarshalIndent(x, "", "    ")
			s = string(b)
		}
		if unquote && strings.HasPrefix(s, `"`) {
			z := ""
			err := json.Unmarshal([]byte(s), &z)
			if err != nil {
				z = s
			}
			return z
		}
		return s
	}

	// Yaml
	if len(s) < 5 {
		b, _ := qutil.Yaml(show.Header)
		return string(b)
	}

	var x interface{}
	json.Unmarshal([]byte(s), &x)
	y, err := qutil.Yaml(&x)
	if err != nil {
		return s
	}

	return string(y)
}
