package error

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	qutil "brocade.be/qtechng/lib/util"

	"github.com/spyzhov/ajson"
)

const maxerrorlines = 100

type ErrorMap map[string]error
type ErrorSlice []error

// NewErrorMap creates a new errormap
func NewErrorMap() ErrorMap {
	return ErrorMap(make(map[string]error))
}

// NewErrorSlice creates a new errorslice
func NewErrorSlice() ErrorSlice {
	return ErrorSlice(make([]error, 0))
}

func (errmap ErrorMap) MarshalJSON() (js []byte, err error) {
	if len(errmap) == 0 {
		return
	}
	errs := make(map[string]map[string]string)
	errs["ERROR"] = make(map[string]string)
	for key, value := range errmap {
		if value == nil {
			continue
		}
		errs["ERROR"][key] = value.Error()
	}
	if len(errs["ERROR"]) == 0 {
		return
	}
	return json.MarshalIndent(errs, "", "    ")
}

func (errslice ErrorSlice) MarshalJSON() (js []byte, err error) {
	if len(errslice) == 0 {
		return
	}
	eslice := FlattenErrors(errslice)
	if len(eslice) == 0 {
		return
	}
	return json.MarshalIndent(eslice, "", "    ")
}

func (errmap ErrorMap) Error() string {
	js, _ := json.Marshal(errmap)
	return string(js)
}

func (errslice ErrorSlice) Error() string {
	eslice := FlattenErrors(errslice)
	if len(eslice) == 0 {
		return "nil"
	}
	blob, _ := json.MarshalIndent(eslice, "", "    ")
	return string(blob)
}

func (errmap *ErrorMap) AddError(key string, err error) {
	if len(*errmap) == 0 {
		x := ErrorMap(make(map[string]error))
		errmap = &x
	}
	(*errmap)[key] = err
}

func (errslice *ErrorSlice) AddError(err error) {
	if len(*errslice) == 0 {
		x := ErrorSlice(make([]error, 0))
		errslice = &x
	}
	*errslice = append(*errslice, err)
}

type QError struct {
	Type    string   `json:"type"`
	Ref     []string `json:"ref"`
	Version string   `json:"version"`
	Project string   `json:"project"`
	File    string   `json:"file"`
	Url     string   `json:"fileurl"`
	Lineno  int      `json:"lineno"`
	Object  string   `json:"object"`

	Msg []string `json:"message"`
}

func (qerr *QError) MarshalJSON() ([]byte, error) {
	return []byte((*qerr).String()), nil
}

type QER struct {
	System string `json:"system"`
}

func (qer *QER) Error() string {
	return qer.System
}

func (qer *QER) String() string {
	return qer.System
}

func (qer *QER) MarshalJSON() ([]byte, error) {
	return []byte(qer.String()), nil
}

func FillQError(qerr *QError) (err *QError) {
	err = new(QError)
	*err = *qerr
	if err.Type == "" {
		err.Type = "ERROR"
	}
	return
}

// func (e *QError) String() (result string) {
// 	r, _ := json.MarshalIndent(e, "", "    ")
// 	result = string(r)
// 	return
// }

// func (e *QError) Error() string {
// 	return e.String()
// }

func (e QError) String() (result string) {
	if e.Url == "" && e.File != "" {
		e.Url = qutil.FileURL(e.File, e.Lineno)
	}
	r, _ := json.MarshalIndent(e, "", "    ")
	result = string(r)
	return
}

func (e QError) Error() string {
	return e.String()
}

func QErrorTune(e error, additional *QError) *QError {
	switch v := e.(type) {
	case *QError:
		if v.Version == "" {
			v.Version = additional.Version
		}
		if v.Project == "" {
			v.Project = additional.Project
		}
		if v.File == "" {
			v.File = additional.File
		}
		if len(additional.Ref) > 0 {
			v.Ref = append(additional.Ref, v.Ref...)
		}
		if len(additional.Msg) > 0 {
			v.Msg = append(additional.Msg, v.Msg...)
		}
		return v
	default:
		additional.Msg = append(additional.Msg, v.Error())
		return additional
	}
}

// ShowError show an error
func ShowError(e error) string {
	host, _ := os.Hostname()

	h := time.Now()
	t := h.Format(time.RFC3339)

	type showerror struct {
		Host  string      `json:"host"`
		Time  string      `json:"time"`
		Error interface{} `json:"ERROR"`
	}

	s := make([]byte, 0)
	switch err := e.(type) {
	case QError, *QError, ErrorMap, ErrorSlice:
		se := showerror{host, t, err}
		s, _ = json.MarshalIndent(se, "", "    ")
	case error:
		se := showerror{host, t, err.Error()}
		s, _ = json.MarshalIndent(se, "", "    ")
	case fmt.Stringer:
		se := showerror{host, t, err.String()}
		s, _ = json.MarshalIndent(se, "", "    ")
	default:
		se := showerror{host, t, e}
		s, _ = json.MarshalIndent(se, "", "    ")
	}

	return string(s)

}

// FlattenErrors
func FlattenErrors(err ErrorSlice) []error {
	if len(err) == 0 {
		return nil
	}
	errs := make([]error, 0)

	for _, e := range err {
		if e == nil {
			continue
		}
		switch v := e.(type) {
		case ErrorSlice:
			if len(v) == 0 {
				continue
			}
			for _, val := range FlattenErrors(v) {
				if val == nil {
					continue
				}
				errs = append(errs, val)
			}
		default:
			errs = append(errs, v)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// ShowResult toont het resultaat in JSON
func ShowResult(r interface{}, jsonpath string, e interface{}, yaml bool) string {
	host, _ := os.Hostname()
	h := time.Now()
	t := h.Format(time.RFC3339)

	type showresult struct {
		Host   string      `json:"host"`
		Time   string      `json:"time"`
		Error  interface{} `json:"ERROR"`
		Result interface{} `json:"RESULT"`
	}
	type jshowresult struct {
		Host   string      `json:"host"`
		Time   string      `json:"time"`
		Jerror string      `json:"jsonpatherror"`
		Error  interface{} `json:"ERROR"`
		Result interface{} `json:"RESULT"`
	}

	se := showresult{}
	switch err := e.(type) {
	case *QError, ErrorMap:
		se = showresult{host, t, err, r}
	case ErrorSlice:
		ers := FlattenErrors(err)
		if len(ers) == 0 {
			ers = nil
		}
		se = showresult{host, t, ers, r}
	case []error:
		ers := FlattenErrors(ErrorSlice(err))
		if len(ers) == 0 {
			ers = nil
		}
		se = showresult{host, t, ers, r}
	case error:
		se = showresult{host, t, err.Error(), r}
	case fmt.Stringer:
		se = showresult{host, t, err.String(), r}
	default:
		se = showresult{host, t, e, r}
	}

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
				if yaml {
					r, _ := qutil.Transform(s, "", yaml)
					return r
				}
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
			se := jshowresult{}
			switch err := e.(type) {
			case *QError, ErrorMap, ErrorSlice:
				se = jshowresult{host, t, ermsg, err, r}
			case error:
				se = jshowresult{host, t, ermsg, err.Error(), r}
			case fmt.Stringer:
				se = jshowresult{host, t, ermsg, err.String(), r}
			default:
				se = jshowresult{host, t, ermsg, e, r}
			}
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
