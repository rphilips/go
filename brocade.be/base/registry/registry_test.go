package registry

import "testing"

func TestRegistry(t *testing.T) {

	testvalue, ok := Registry["qtechng-test"]
	if !ok {
		t.Errorf("Test 'qtechng-test' does not exists in Registry")
	} else if testvalue != "test-entry" {
		t.Errorf("'qtechng-test' leads to wrong value")
	}
}
