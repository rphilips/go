package gmail

import (
	"testing"

	"brocade.be/base/fs"
)

func TestSend(t *testing.T) {

	to := []string{"richard.philips@gmail.com", "ria.vercruysse@gmail.com"}
	subject := "Uitslagen van dit weekend!"
	text := ""

	html, err := fs.Fetch("/home/rphilips/Dropbox/Chess/IC/PK/2022-2023/R3.html")
	if err != nil {
		t.Errorf("Error %s", err)
	}
	err = Send(to, nil, nil, subject, text, string(html), nil)
	if err != nil {
		t.Errorf("Error %s", err)
	}

}
