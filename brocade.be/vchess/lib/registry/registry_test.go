package registry

import "testing"

func TestRegistry(t *testing.T) {

	testvalue, ok := Registry["test"]
	if !ok {
		t.Errorf("Test 'test' does not exists in Registry: %v", Registry)
	} else if testvalue.(string) != "vchess-test" {
		t.Errorf("'test' leads to wrong value")
	}
}
