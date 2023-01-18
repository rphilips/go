package gmail

import (
	"encoding/json"
	"fmt"
	"net/smtp"
	"path/filepath"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qemail "github.com/jordan-wright/email"
)

func Send(to []string, cc []string, bcc []string, subject string, text string, html string, attachments []string, dir string) (err error) {

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
	if err != nil || dir == "" {
		return err
	}

	err = archive(to, cc, bcc, subject, text, html, attachments, time.Now(), dir)

	return err

}

func archive(to []string, cc []string, bcc []string, subject string, text string, html string, attachments []string, t time.Time, dir string) (err error) {

	type arc struct {
		To          []string  `json:"to"`
		Cc          []string  `json:"cc"`
		Bcc         []string  `json:"bcc"`
		Subject     string    `json:"subject"`
		Text        string    `json:"text"`
		HTML        string    `json:"html"`
		Attachments []string  `json:"attachments"`
		Time        time.Time `json:"time"`
	}

	backup := arc{
		To:          to,
		Cc:          cc,
		Bcc:         bcc,
		Subject:     subject,
		Text:        text,
		HTML:        html,
		Attachments: attachments,
		Time:        t,
	}
	data, err := json.MarshalIndent(backup, "", "    ")
	if err != nil {
		return
	}
	people := make(map[string]bool)

	for _, p := range to {
		if strings.Contains(p, "@") {
			people[p] = true
		}
	}
	for _, p := range cc {
		if strings.Contains(p, "@") {
			people[p] = true
		}
	}
	for _, p := range bcc {
		if strings.Contains(p, "@") {
			people[p] = true
		}
	}

	ts := t.String()
	d, a, ok := strings.Cut(ts, " ")
	if !ok {
		d = ts
		a = "x"
	}
	for p := range people {

		adir := filepath.Join(dir, p, d)
		qfs.MkdirAll(adir, "process")
		fname := filepath.Join(adir, a)
		fmt.Println(fname)
		qfs.Store(fname, data, "")
	}
	return nil
}
