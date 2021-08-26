package error

import (
	"encoding/json"
	"strconv"
	"strings"

	qutil "brocade.be/qtechng/lib/util"
)

type QError struct {
	Type    string   `json:"type"`
	Ref     []string `json:"ref"`
	Version string   `json:"version"`
	Project string   `json:"project"`
	QPath   string   `json:"qpath"`
	File    string   `json:"file"`
	Url     string   `json:"fileurl"`
	Lineno  int      `json:"lineno"`
	Object  string   `json:"object"`

	Msg []string `json:"message"`
}

func (qerr QError) Error() string {
	return qerr.String()
}

func (pqerr *QError) String() (result string) {
	r, _ := json.Marshal(pqerr)
	result = string(r)
	return
}

func (qerr QError) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{})
	if len(qerr.Ref) == 0 {
		m["ref"] = "unknown"
	} else {
		m["ref"] = strings.Join(qerr.Ref, " ; ")
	}
	if qerr.Version != "" {
		m["version"] = qerr.Version
	}
	if qerr.Project != "" {
		m["project"] = qerr.Project
	}
	if qerr.QPath != "" {
		m["qpath"] = qerr.QPath
	}
	if qerr.File != "" {
		m["file"] = qerr.File
	}

	if qerr.Url == "" && qerr.File != "" {
		qerr.Url = qutil.FileURL(qerr.File, qerr.QPath, qerr.Lineno)
	}

	if qerr.Url != "" {
		m["fileurl"] = qerr.Url
	} else {
		m["fileurl"] = ""
	}
	if qerr.Lineno > 0 {
		m["lineno"] = strconv.Itoa(qerr.Lineno)
		lineno := m["lineno"].(string)
		fileurl := m["fileurl"].(string)

		if fileurl != "" && !strings.HasSuffix(fileurl, "#"+lineno) {
			m["fileurl"] = fileurl + "#" + lineno
		}

	}
	if qerr.Object != "" {
		m["object"] = qerr.Object
	}
	if len(qerr.Msg) != 0 {
		message := make([]interface{}, 0)
		for _, msg := range qerr.Msg {
			msg = strings.TrimSpace(msg)
			if strings.TrimSpace(msg) == "" {
				continue
			}
			if !strings.HasPrefix(msg, "{") {
				message = append(message, msg)
				continue
			}
			mmsg := make(map[string]interface{})
			e := json.Unmarshal([]byte(msg), &mmsg)
			if e != nil {
				message = append(message, msg)
				continue
			}
			mers, ok := mmsg["error"]
			if ok {
				lmers := make([]string, 0)
				e := json.Unmarshal([]byte(mers.(string)), &lmers)
				if e == nil {
					mmsg["error"] = lmers
				}
			}

			message = append(message, mmsg)
		}
		m["message"] = message
	}

	return json.MarshalIndent(m, "", "    ")
}

type ErrorSlice []error

// NewErrorSlice creates a new errorslice
func NewErrorSlice() ErrorSlice {
	return ErrorSlice(make([]error, 0))
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

func (errslice ErrorSlice) Error() string {
	eslice := FlattenErrors(errslice)
	if len(eslice) == 0 {
		return "nil"
	}
	blob, _ := json.MarshalIndent(eslice, "", "    ")
	return string(blob)
}

func (errslice *ErrorSlice) AddError(err error) {
	if len(*errslice) == 0 {
		x := ErrorSlice(make([]error, 0))
		errslice = &x
	}
	*errslice = append(*errslice, err)
}

func FillQError(qerr *QError) (err *QError) {
	err = new(QError)
	*err = *qerr
	if err.Type == "" {
		err.Type = "ERROR"
	}
	return
}

func QErrorTune(e error, additional *QError) *QError {
	if e == nil {
		return additional
	}
	switch v := e.(type) {
	case QError:
		e = &v
	}

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
		if additional.QPath != "" && v.QPath == "" {
			v.QPath = additional.QPath
		}
		if v.QPath != "" && v.File == v.QPath {
			v.File = ""
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

func FlattenErrors(err interface{}) []error {
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
	case ErrorSlice:
		if len(v) == 0 {
			return nil
		}
		errs = append(errs, v...)
	case *ErrorSlice:
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
	case map[string]interface{}:
		if len(v) == 0 {
			return nil
		}
		for _, e := range v {
			errs = append(errs, e.(error))
		}
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
		case QError:
			errs2 = append(errs2, v)
		case *QError:
			errs2 = append(errs2, *v)
		case ErrorSlice:
			if len(v) == 0 {
				continue
			}
			es := []error(v)
			errs2 = append(errs2, FlattenErrors(es)...)
		case *ErrorSlice:
			if len(*v) == 0 {
				continue
			}
			es := []error(*v)
			errs2 = append(errs2, FlattenErrors(es)...)
		case error:
			if v == nil {
				continue
			}
			es := QError{
				Msg: []string{v.Error()},
			}
			errs2 = append(errs2, es)
		default:
			errs2 = append(errs2, v)
		}
	}
	if len(errs2) == 0 {
		return nil
	}
	return errs2
}

func ErrorMsg(e error) []string {

	if e == nil {
		return nil
	}
	switch v := e.(type) {
	case QError:
		e = &v
	}

	switch v := e.(type) {
	case *QError:
		if len(v.Msg) == 0 {
			return nil
		}
		return v.Msg
	default:
		return []string{e.Error()}
	}
}

func ExtractEMsg(e error, fname string, blob []byte) (msg []string, lineno int) {
	msg = make([]string, 0)
	if e == nil {
		return
	}
	switch v := e.(type) {
	case *QError:
		lineno = v.Lineno
		msg = v.Msg
	case QError:
		lineno = v.Lineno
		msg = v.Msg
	default:
		m := qutil.ExtractMsg(e.Error(), fname)
		no, line := qutil.ExtractLineno(m, blob)
		lineno = no
		msg = []string{m + " :: " + line}
	}
	return
}
