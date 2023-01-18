package gmail

import (
	"testing"

	"brocade.be/base/fs"
)

func TestSend(t *testing.T) {

	to := []string{"richard.philips@gmail.com"}
	subject := "Uitslagen2 van dit weekend!"
	text := ""

	html, err := fs.Fetch("/home/rphilips/Dropbox/Chess/IC/PK/2022-2023/R3.html")
	if err != nil {
		t.Errorf("Error %s", err)
	}
	err = Send(to, nil, nil, subject, text, string(html), nil, "/home/rphilips/tmp")
	if err == nil {
		t.Errorf("Error1 %s", err)
	}

}
