package time

import (
	"testing"
)

func TestNow(t *testing.T) {
	h := Now()
	t.Errorf("Found: [%s]", h)

}
