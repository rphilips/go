package result

import (
	"encoding/json"
	"os"
	"strings"
	"time"

	"github.com/spyzhov/ajson"
)

// ShowResult toont het resultaat in JSON
func ShowResult(r interface{}, jsonpath string) string {
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
					return string(blob)
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
					return strings.Join(r, "\n")
				}
			}
		}
		if ermsg != "" {
			se := jshowresult{host, t, ermsg, r}
			s, _ = json.MarshalIndent(se, "", "    ")
			return string(s)
		}
	}

	return string(s)
}

// Show shows combined output
func Show(stdout string, stderr string) string {
	if stderr != "" {
		return stderr
	}
	return stdout
}
