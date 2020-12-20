package xfile

import (
	"strings"
	"testing"
)

func TestX401(t *testing.T) {

	body := `Hello x4_varruntime(UDcaCode) World`
	widget := makewidget("screen", body)
	expect := `-:Hello |X:w $$Runtime^uwwwscr("UDcaCode","")|-: World`

	result, err := widget.Resolve()
	found := strings.Join(result, "|")

	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if strings.Join(result, "|") != expect {
		t.Errorf("Error: [%s]\n\nExpected: [%s]\nFound   : [%s]\n", body, expect, found)
	}

}

func TestX402(t *testing.T) {

	body := `Hello x4_ x4_ World`
	widget := makewidget("screen", body)
	expect := `-:Hello x4_ x4_ World`

	result, err := widget.Resolve()
	found := strings.Join(result, "|")

	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	if strings.Join(result, "|") != expect {
		t.Errorf("Error: [%s]\n\nExpected: [%s]\nFound   : [%s]\n", body, expect, found)
	}

}

func TestX403(t *testing.T) {

	body := `Hello World
x4_if(.END_1,$G(FDedit))
A
B
.END_1
rest`
	widget := makewidget("screen", body)
	expect := `-:Hello World#|I:4:$G(FDedit)|-:#A#B|-:|-:#rest`
	expect = strings.ReplaceAll(expect, "|", "\n")

	result, err := widget.Resolve()

	if err != nil {
		t.Errorf("Error: %s", err)
		return
	}
	sresult := strings.Join(result, "|")
	sresult = strings.ReplaceAll(sresult, "\n", "#")
	sresult = strings.ReplaceAll(sresult, "|", "\n")
	if sresult != expect {
		t.Errorf("Error: [%s]\n\nExpected: \n[%s]\nFound   : \n[%s]\n", body, expect, sresult)
	}

}

func makewidget(ty string, body string) *Widget {
	widget := Widget{
		ID:      ty + " " + "myWidget",
		Body:    body,
		Line:    "85",
		Version: "6.00",
		Source:  "/project/myxfile.x",
	}
	return &widget
}
