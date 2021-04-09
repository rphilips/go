package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spyzhov/ajson"
	qyaml "gopkg.in/yaml.v2"
)

func JSONpath(b []byte, jsonpath string) (string, error) {
	result, err := ajson.JSONPath(b, jsonpath)
	if err != nil {
		return string(b), err
	}

	switch len(result) {
	case 0:
		return "", nil
	case 1:
		blob, _ := ajson.Marshal(result[0])
		return string(blob), nil
	default:
		r := make([]string, len(result)+2)
		r[0] = "["
		for i, pnode := range result {
			blob, _ := ajson.Marshal(pnode)
			comma := ","
			if i == len(result)-1 {
				comma = ""
			}
			s := "    " + string(blob) + comma
			r[i+1] = s
			continue
		}
		r[len(result)+1] = "]"
		res := strings.Join(r, "\n")
		return res, nil
	}
}

func ParsePath(jsonpath string) error {
	_, err := ajson.ParseJSONPath(jsonpath)
	return err
}

func Yaml(in interface{}) ([]byte, error) {
	return qyaml.Marshal(in)
}

func Transform(input []byte, jsonpath string, yaml bool) (output string, err error) {
	if jsonpath == "" && !yaml {
		return string(input), nil
	}
	if jsonpath != "" {
		_, err = ajson.ParseJSONPath(jsonpath)
		if err != nil {
			output = string(input)
			return
		}
		if input == nil {
			return
		}
		result, err := ajson.JSONPath(input, jsonpath)
		if err != nil {
			return string(input), err
		}
		output = fmt.Sprint(result)
	} else {
		output = string(input)
	}
	if yaml {
		var x interface{}
		json.Unmarshal([]byte(output), &x)
		y, e := qyaml.Marshal(&x)
		if e == nil {
			output = string(y)
		}
		if len(output) < 10 && strings.TrimSpace(output) == "" {
			json.Unmarshal(input, &x)
			y, e := qyaml.Marshal(&x)
			if e == nil {
				output = string(y)
			}
		}
	}
	return
}

func Encode(s string) string {
	z, _ := json.Marshal(s)
	return string(z)
}
