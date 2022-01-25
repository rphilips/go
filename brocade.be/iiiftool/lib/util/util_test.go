package util

import (
	"fmt"
	"testing"
)

func TestStrReverse(t *testing.T) {
	test := " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
	expected := "~}|{zyxwvutsrqponmlkjihgfedcba`_^]\\[ZYXWVUTSRQPONMLKJIHGFEDCBA@?>=<;:9876543210/.-,+*)('&%$#\"! "
	result := StrReverse(test)
	if result != expected {
		t.Errorf(fmt.Sprintf("\nResult: \n[%s]\nExpected: \n[%s]\n", result, expected))
	}

}
