package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	bstrings "brocade.be/base/strings"
	btime "brocade.be/base/time"
	plog "brocade.be/pbladng/lib/log"
	pnext "brocade.be/pbladng/lib/next"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
)

var warnCmd = &cobra.Command{
	Use:   "warn",
	Short: "Warn `pblad`",
	Long:  "Warn `pblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `pblad warn myfile.pb`,
	RunE:    warn,
}

func init() {
	rootCmd.AddCommand(warnCmd)
}

func warn(cmd *cobra.Command, args []string) error {

	year, week, mailed, _ := pstructure.DocRef("")
	id := fmt.Sprintf("%d-%02d", year, week)

	logfile := plog.Logfile()
	info := make(map[string]string)
	info["logfile"] = logfile
	info["id"] = id

	log, err := plog.Fetch()
	if err != nil {
		if Fdebug {
			fmt.Println("debug: invalid-logfile")
		}
		logmail("invalid-logfile", info, log)
		return err
	}

	// timestamp
	now := time.Now()
	log["crontab"] = btime.StringTime(&now, "I")
	plog.Store(log)

	// check last new
	nextid, date := pnext.NextToNew(id)
	maildate := pnext.MailDate(id)

	if maildate <= btime.StringDate(&now, "I") {
		_, stamp := plog.GetMark("new-warn")
		if stamp == "" || stamp[:10] < btime.StringDate(&now, "I") {
			info["nextid"] = nextid
			info["date"] = date
			then := btime.DetectDate(date)
			diff := then.Sub(now)
			days := int(diff.Hours() / 24)
			info["days"] = strconv.Itoa(days)
			logmail("new", info, log)
			plog.SetMark("new-warn", "")
		}
	}

	// check last distribute
	value, _ := plog.GetMark("distribute")
	if Fdebug {
		fmt.Println("debug: " + bstrings.JSON(map[string]string{"id": id, "nextid": nextid, "maildate": maildate, "now": btime.StringDate(&now, "I"), "value": value}))
	}
	if value != id {
		_, stamp := plog.GetMark("distribute-warn")
		if stamp == "" || stamp[:10] < btime.StringDate(&now, "I") {
			info["id"] = id
			logmail("distribute", info, log)
			plog.SetMark("distribute-warn", "")
		}
	}

	// check mailed
	maildate = pnext.MailDate(id)
	then := btime.DetectDate(maildate)
	if mailed == "" && then.Before(now.AddDate(0, 0, 2)) {
		_, stamp := plog.GetMark("mailed-warn")
		if stamp == "" || stamp[:10] < btime.StringDate(&now, "I") {
			info["id"] = id
			info["until"] = btime.StringDate(then, "D")
			logmail("mailed", info, log)
			plog.SetMark("mailed-warn", "")
		}

	}

	return nil
}

func logmail(logid string, info map[string]string, log map[string]string) bool {
	if Fdebug {
		fmt.Println("debug1: ", logid)
	}

	now := time.Now()
	mailid := "mail-" + logid
	t := log[mailid]
	if t != "" && btime.StringDate(&now, "I") == t {
		return false
	}
	t = btime.StringDate(&now, "I")

	warnprops := pregistry.Registry["warn"]

	x := warnprops.(map[string]any)
	mail := x[logid].(map[string]any)
	subject := mail["subject"].(string)
	body := mail["body"].(string)

	warnmail := x["mail"].(string)
	if warnmail == "" {
		return false
	}

	for k, v := range info {
		subject = strings.ReplaceAll(subject, "{"+k+"}", v)
		body = strings.ReplaceAll(body, "{"+k+"}", v)
	}
	subject = "[pblad] " + subject

	mails := []string{warnmail}
	mdir, ok := pregistry.Registry["mail-dir"]
	maildir := ""
	if ok {
		maildir = mdir.(string)
	}
	if Fdebug {
		fmt.Println("debug2: ", logid)
	} else {
		bmail.Send(mails, nil, nil, subject, body, "", nil, maildir)
		log[mailid] = t
		bfs.Store(info["logfile"], log, "process")
	}
	return true

}
