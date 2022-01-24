package cmd

import (
	"encoding/json"
	"testing"
)

func TestFormat(t *testing.T) {

	s1 := "Hello World"
	r, _ := json.MarshalIndent(s1, "", "    ")
	t.Errorf("\n\nObject: %v\nType: %T\nResult: %s\n\n", s1, s1, r)

	var s2 float64 = 123.5
	r, _ = json.MarshalIndent(s2, "", "    ")
	t.Errorf("\n\nObject: %v\nType: %T\nResult: %s\n\n", s2, s2, r)

	s3 := []string{"a", "b"}
	r, _ = json.MarshalIndent(s3, "", "    ")
	t.Errorf("\n\nObject: %v\nType: %T\nResult: \n%s\n\n", s3, s3, r)

	s4 := map[string]string{"a": "A", "b": "B"}
	r, _ = json.MarshalIndent(s4, "", "    ")
	t.Errorf("\n\nObject: %v\nType: %T\nResult: \n%s\n\n", s4, s4, r)
}
