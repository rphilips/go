package holy

import (
	"strings"
	"testing"

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
