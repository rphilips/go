package pool

import (
	"strconv"
	"testing"
)

func double(si string) (interface{}, error) {
	i, _ := strconv.Atoi(si)

	return interface{}(2 * i), nil
}

func TestNumber(t *testing.T) {
	keys := make([]string, 0)
	for i := 0; i < 1000; i++ {
		keys = append(keys, strconv.Itoa(i))
	}
	result := PMap(keys, double)

	if len(result) != len(keys) {
		t.Errorf("TestNumber failed")
	}

	return
}

func TestNumberSelect(t *testing.T) {
	keys := make([]int, 0)
	for i := 0; i < 10000; i++ {
		keys = append(keys, 2*i)
	}
	fn := func(k int) bool {
		return keys[k] < 20
	}
	result := NSelect(len(keys), fn)

	if len(result) != 10 {
		t.Errorf("Test failed: %v", result)
	}

	return
}
