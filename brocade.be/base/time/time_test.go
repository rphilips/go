package time

import (
	"strconv"
	"testing"
	"time"
)

func TestNow1(t *testing.T) {
	h := Now()
	t.Errorf("Found: [%s]", h)

}

func TestNow2(t *testing.T) {
	now := time.Now()
	year := strconv.Itoa(now.Year())
	times := []string{
		"15/03/" + year,
		"15/03",
		"15 maa",
		"15 maa " + year,
		"15-03-" + year,
	}
	for _, s := range times {
		tim := DetectDate(s)
		if tim == nil || tim.Day() != 15 || tim.Month() != 3 || tim.Year() != now.Year() {
			t.Errorf("Found: %v", tim)
		}
	}
}
