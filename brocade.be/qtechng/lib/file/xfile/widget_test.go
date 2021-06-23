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

func TestY(t *testing.T) {
	body := `screen meta:
	<table class="metascreen-table">
	x4_if(.END_1,FDid'="")
	<tr>
	<td class="metascreen-key">x4_varcode(identificatie)</td>
	<td class="metascreen-id">x4_varruntime(FDid)
	<input type="hidden" name="FDid" value="x4_varruntime(FDid)">
	<input type="hidden" name="UDsrchIt" value="x4_varruntime(FDid)">
	<span class="metascreen-lookup">m4_lookupCopy('x4_varruntime(FDid)')</span>
.END_1
	x4_if(.END_2,FDid="")</td></tr>
	
	<tr>
	<td class="metascreen-key">x4_varcode(identificatie)</td>
	<td>
	<input type="text" name=FDid value="" size=42 ><input type="hidden" name="UDsrchIt" value="">
	.END_2</td></tr>
	<tr>
	
	
	m4_documentElementRootEnd`
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

func TestZ(t *testing.T) {
	body := `format meta:
	<table class="metascreen-table">
	x4_if(.END_1,x4_parconstant(1)'="")
	<span class="metascreen-lookup">m4_lookupCopy('x4_varruntime(FDid)')</span>
.END_1`
	widget := makewidget("format", body)
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
