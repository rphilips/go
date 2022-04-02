package docman

import (
	"io"
	"testing"

	qfs "brocade.be/base/fs"
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

func TestS1(t *testing.T) {
	ids := "/uakaga/104722/1.pdf"
	id := DocmanID(ids)
	calc := id.Location()
	expect := ""
	if calc != expect {
		t.Errorf("id %s\ncalc   %s\nexpect %s", id, calc, expect)
		return
	}

	reader, err := id.Reader()
	if reader != nil {
		x, _ := io.ReadAll(reader)
		qfs.Store("/home/rphilips/Desktop/test.pdf", x, "qtech")
		t.Errorf("found: %v", x)
	}
	if err != nil {
		if err.Error() != "cannot find docman `/uakaga/104722/1.pdf`" {
			t.Errorf("err: %v", err)
			return
		}
	}

}
