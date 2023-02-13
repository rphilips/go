package tools

import (
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	params := []string{"pblad", "{verb}"}
	keys := map[string]string{"verb": "about"}

	output, err := Launch(params, keys, "", true, false)

	if err != nil {
		t.Errorf("Problem:\noutput:`%s`\nerror:`%s`\n", output, err)
	}

	if err == nil && !strings.Contains(string(output), "REGISTRY") {
		t.Errorf("Problem:\noutput:`%s`\n", output)
	}

}
