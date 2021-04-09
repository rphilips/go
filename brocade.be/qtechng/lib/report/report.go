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
	Header header      `json:"aHEADER" yaml:"aHEADER"`
	Result interface{} `json:"bRESULT" yaml:"bRESULT"`
	Errors []error     `json:"cERRORS" yaml:"cERRORS"`
}

func Report(r interface{}, e interface{}, jsonpath string, yaml bool) string {

	show := report{}

	// header

	host, _ := os.Hostname()
	h := header{
		Host: host,
		Time: time.Now().Format(time.RFC3339),
		Args: os.Args,
	}
	if jsonpath != "" {
		err := qutil.ParsePath(jsonpath)
		if err != nil {
			h.Jerror = err.Error()
			jsonpath = ""
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
	if jsonpath != "" {
		s, _ = qutil.JSONpath(b, jsonpath)
	}
	s = strings.TrimSpace(s)

	if !yaml {
		var x interface{}
		err := json.Unmarshal([]byte(s), &x)
		if err == nil {
			b, _ := json.MarshalIndent(x, "", "    ")
			return string(b)
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
