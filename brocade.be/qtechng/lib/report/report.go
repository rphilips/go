package report

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
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
	Errors interface{} `json:"ERRORS" yaml:"ERRORS"`
}

func Report(r interface{}, e interface{}, jsonpath []string, yaml bool, unquote bool, joiner string, silent bool, file string) string {
	if silent {
		return ""
	}
	show := report{}
	withfile := func(s string) string {
		if silent {
			return ""
		}
		if file == "" {
			return s
		}
		qfs.Store(file, s, "qtech")
		return s
	}

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

	errs := qerror.FlattenErrors(e)
	if len(errs) == 0 {
		show.Errors = nil
	} else {
		b, e := json.Marshal(errs)
		if e == nil {
			var i []interface{}
			json.Unmarshal(b, &i)
			show.Errors = qutil.FlattenInterface(i)
			switch v := show.Errors.(type) {
			case []interface{}:
				if len(v) == 0 {
					show.Errors = nil
				}
			default:
				show.Errors = []interface{}{v}
			}
		} else {
			show.Errors = errs
		}
	}
	// to JSON

	b, _ := json.MarshalIndent(show, "", "    ")
	if show.Errors != nil {
		return withfile(string(b))
	}

	if silent {
		return ""
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
		if unquote {
			switch {
			case strings.HasPrefix(s, `"`):
				z := ""
				err := json.Unmarshal([]byte(s), &z)
				if err != nil {
					z = s
				}
				return withfile(z)
			case strings.HasPrefix(s, `[`):
				z := make([]string, 0)
				err := json.Unmarshal([]byte(s), &z)
				if err != nil {
					return withfile(s)
				}
				return withfile(strings.Join(z, qutil.Joiner(joiner)))
			}
		}

		return withfile(s)
	}

	// Yaml
	if len(s) < 5 {
		b, _ := qutil.Yaml(show.Header)
		return withfile(string(b))
	}

	var x interface{}
	json.Unmarshal([]byte(s), &x)
	y, err := qutil.Yaml(&x)
	if err != nil {
		return withfile(s)
	}

	return withfile(string(y))
}
