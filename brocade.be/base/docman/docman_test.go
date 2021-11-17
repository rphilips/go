package docman

import (
	"testing"
)

func TestD1(t *testing.T) {
	ids := "/exchange/d8e365/richard_philips.pdf"
	id := DocmanID(ids)
	db := id.DB()
	expect := "exchange"
	if db != expect {
		t.Errorf("id %s; db %s; expect %s", id, db, expect)
		return
	}
}

func TestD2(t *testing.T) {
	ids := "/misc/fddffc/extrafile.html"
	id := DocmanID(ids)
	calc := id.Location()
	expect := "/library/database/docman/misc/xf/dd/ffcextrafile.html.rm20220313"
	if calc != expect {
		t.Errorf("id %s\ncalc   %s\nexpect %s", id, calc, expect)
		return
	}
}

func TestD3(t *testing.T) {
	ids := "/misc/fedb67/3.jpg"
	id := DocmanID(ids)
	calc := id.Location()
	expect := "/library/database/docman/misc/xf/ed/b673.jpg"
	if calc != expect {
		t.Errorf("id %s\ncalc   %s\nexpect %s", id, calc, expect)
		return
	}
}
