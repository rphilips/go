package util

import (
	"encoding/json"
	"fmt"

	"github.com/spyzhov/ajson"
	qyaml "gopkg.in/yaml.v2"
)

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
	}
	return
}
