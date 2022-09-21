package holy

import (
	"strings"
	"testing"
	"time"

	ptools "brocade.be/pbladng/lib/tools"
)

func TestEastern(t *testing.T) {
	peastern, _, _ := Pasen(2022)
	weekday := ptools.StringDate(peastern, "D")
	if weekday != "zondag 17 april 2022" {
		t.Errorf("Eastern %d is on %s", 2022, weekday)
		return
	}
	pmonday, _, _ := Paasmaandag(2022)
	weekday = ptools.StringDate(pmonday, "D")
	if !strings.HasPrefix(weekday, "maandag") {
		t.Errorf("Eastern monday %d is on %s", 2022, weekday)
		return
	}
	pwoe, _, _ := Aswoensdag(2022)
	weekday = ptools.StringDate(pwoe, "D")
	if !strings.HasPrefix(weekday, "woensdag") {
		t.Errorf("Aswoensdag %d is on %s", 2022, weekday)
		return
	}
}

func TestHoly(t *testing.T) {
	peastern, _, _ := Pasen(2022)
	result := Today(peastern)
	if len(result) != 1 {
		t.Errorf("Only Eastern, found:  %v", result)
		return
	}
	now := time.Now()
	test := time.Date(2022, 10, 2, 0, 0, 0, 0, now.Location())
	if !issunday(&test) {
		t.Errorf("Should be a sunday")
	}
	result = Today(&test)
	if len(result) != 1 || result[0] != "27e ZONDAG DOOR HET JAAR" {
		doop, _, _ := Doopjezus(2022)
		t.Errorf("Only Eastern, found:  %v %s", result, *doop)
	}
}
