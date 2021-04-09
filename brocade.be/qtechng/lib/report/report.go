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
	Jerror string   `json:"jsonpatherror" yaml:"jsonpatherror"`
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
	show.Errors = flatten(e)

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

func flatten(err interface{}) []error {
	if err == nil {
		return nil
	}
	errs := make([]error, 0)
	switch v := err.(type) {
	case []error:
		if len(v) == 0 {
			return nil
		}
		errs = append(errs, v...)
	case qerror.ErrorSlice:
		if len(v) == 0 {
			return nil
		}
		errs = append(errs, v...)
	case *qerror.ErrorSlice:
		es := []error(*v)
		if len(es) != 0 {
			errs = append(errs, es...)
		} else {
			return nil
		}
	case error:
		if v == nil {
			return nil
		}
		errs = append(errs, v)
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		for _, e := range v {
			errs = append(errs, e.(error))
		}
	default:
		errs = append(errs, v.(error))
	}
	errs2 := make([]error, 0)

	for _, e := range errs {
		if e == nil {
			continue
		}
		switch v := e.(type) {
		case qerror.ErrorSlice:
			if len(v) == 0 {
				continue
			}
			es := []error(v)
			errs2 = append(errs2, flatten(es)...)
		case *qerror.ErrorSlice:
			if len(*v) == 0 {
				continue
			}
			es := []error(*v)
			errs2 = append(errs2, flatten(es)...)
		case error, interface{}:
			if v == nil {
				continue
			}
			errs2 = append(errs2, v)
		default:
			errs2 = append(errs2, v)
		}
	}
	if len(errs2) == 0 {
		return nil
	}
	return errs2
}
