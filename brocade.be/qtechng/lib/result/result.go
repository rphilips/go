package result

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spyzhov/ajson"
)

// ShowResult toont het resultaat in JSON
func ShowResult(r interface{}, jsonpath string, yaml bool) string {
	host, _ := os.Hostname()
	h := time.Now()
	t := h.Format(time.RFC3339)

	type showresult struct {
		Host   string      `json:"host"`
		Time   string      `json:"time"`
		Result interface{} `json:"RESULT"`
	}
	type jshowresult struct {
		Host   string      `json:"host"`
		Time   string      `json:"time"`
		Jerror string      `json:"jsonpatherror"`
		Result interface{} `json:"RESULT"`
	}

	se := showresult{host, t, r}

	s, _ := json.MarshalIndent(se, "", "    ")

	if jsonpath != "" {
		result, err := ajson.JSONPath(s, jsonpath)
		ermsg := ""
		if err != nil {
			ermsg = err.Error()
		}
		if err == nil {
			switch len(result) {
			case 0:
				return ""
			case 1:
				blob, err := ajson.Marshal(result[0])
				if err == nil {
					r, _ := qutil.Transform(blob, "", yaml)
					return r
				}
				ermsg = err.Error()
			default:
				r := make([]string, len(result)+2)
				r[0] = "["
				for i, pnode := range result {
					blob, err := ajson.Marshal(pnode)
					if err == nil {
						comma := ","
						if i == len(result)-1 {
							comma = ""
						}
						s := "    " + string(blob) + comma
						r[i+1] = s
						continue
					}
					ermsg = err.Error()
					break
				}
				if ermsg == "" {
					r[len(result)+1] = "]"
					res := strings.Join(r, "\n")
					if !yaml {
						return res
					}
					res, _ = qutil.Transform([]byte(res), "", yaml)
					return res
				}
			}
		}
		if ermsg != "" {
			se := jshowresult{host, t, ermsg, r}
			s, _ = json.MarshalIndent(se, "", "    ")
			if !yaml {
				return string(s)
			}
			res, _ := qutil.Transform(s, "", yaml)
			return res
		}
	}
	if !yaml {
		return string(s)
	}
	res, _ := qutil.Transform(s, "", yaml)
	return res
}

// Show shows combined output
func Show(stdout string, stderr string) string {
	if stderr != "" {
		return stderr
	}
	return stdout
}
