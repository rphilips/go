package action

import (
	"encoding/csv"
	"fmt"
	"os"

	qutil "brocade.be/goyo/lib/util"
	qyottadb "brocade.be/goyo/lib/yottadb"
)

func ZWR(text string) []string {
	if text == "" {
		return nil
	}
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	show := make(chan qyottadb.VarReport, 100)
	go qyottadb.ZWR(gloref2, show, "", true)
	for report := range show {
		if report.Err != nil {
			qutil.Error(report.Err)
			break
		}
		if report.Gloref == "" {
			break
		}
		fmt.Println(report.Gloref + "=" + report.Value)
	}
	h := []string{"zwr " + gloref}
	if gloref2 != gloref {
		h = append(h, "zwr "+gloref2)
	}
	return h
}

func CSV(text string, output string) []string {
	if text == "" {
		return nil
	}
	w := csv.NewWriter(os.Stdout)
	if output != "" {
		f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		defer f.Close()
		w = csv.NewWriter(f)
	}
	gloref, _ := SplitRefValue(text)
	gloref2 := qyottadb.N(gloref)
	show := make(chan qyottadb.SubReport, 100)
	go qyottadb.CSV(gloref2, show, "", true)
	for report := range show {
		if report.Err != nil {
			qutil.Error(report.Err)
			break
		}
		if len(report.Subs) == 0 {
			break
		}
		w.Write(append(report.Subs, report.Value))
	}
	w.Flush()
	h := []string{"csv " + gloref}
	if gloref2 != gloref {
		h = append(h, "csv "+gloref2)
	}
	return h
}
