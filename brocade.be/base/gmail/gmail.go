package gmail

import (
	"net/smtp"
	"strings"

	qregistry "brocade.be/base/registry"
	qemail "github.com/jordan-wright/email"
)

func Send(to []string, cc []string, bcc []string, subject string, text string, html string, attachments []string) (err error) {

	from := qregistry.Registry["gmail-sender"] // "Jordan Wright <test@gmail.com>"
	frompart := from
	if strings.Contains(frompart, "<") && strings.Contains(frompart, ">") {
		_, frompart, _ = strings.Cut(frompart, "<")
		frompart, _, _ = strings.Cut(frompart, ">")
	}
	frompart = strings.TrimSpace(frompart)
	password := qregistry.Registry["gmail-password"]
	smtpHost := qregistry.Registry["gmail-smtpserver"] // "smtp.gmail.com:587"
	smtpPort := qregistry.Registry["gmail-smtpport"]   // "587"

	e := qemail.NewEmail()
	e.From = from
	e.To = to
	e.Bcc = bcc
	e.Cc = cc
	e.Subject = subject
	e.Text = []byte(text)
	e.HTML = []byte(html)

	for _, attachment := range attachments {
		e.AttachFile(attachment)
	}

	err = e.Send(smtpHost+":"+smtpPort, smtp.PlainAuth("", frompart, password, smtpHost))

	return err

}
