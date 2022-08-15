package tools

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func StringDate(t *time.Time, mode string) string {
	mode = strings.TrimSpace(strings.ToUpper(mode))
	if mode == "" {
		mode = "I"
	}
	if strings.HasPrefix(mode, "I") {
		return fmt.Sprintf("%v", t)[:10]
	}
	imonth := t.Month()
	month := ""
	switch imonth {
	case 1:
		month = "januari"
	case 2:
		month = "februari"
	case 3:
		month = "maart"
	case 4:
		month = "april"
	case 5:
		month = "mei"
	case 6:
		month = "juni"
	case 7:
		month = "juli"
	case 8:
		month = "augustus"
	case 9:
		month = "september"
	case 10:
		month = "oktober"
	case 11:
		month = "november"
	case 12:
		month = "december"
	}
	if !strings.HasPrefix(mode, "D") {
		return fmt.Sprintf("%d %s %d", t.Day(), month, t.Year())
	}
	weekday := t.Weekday().String()
	day := weekday

	switch weekday {
	case "Sunday":
		day = "zondag"
	case "Monday":
		day = "maandag"
	case "Tuesday":
		day = "dinsdag"
	case "Wednesday":
		day = "woensdag"
	case "Thursday":
		day = "donderdag"
	case "Friday":
		day = "vrijdag"
	case "Saturday":
		day = "zaterdag"
	}

	return fmt.Sprintf("%s %d %s %d", day, t.Day(), month, t.Year())
}

func NewDate(date string) (t *time.Time, after string, err error) {

	udate := strings.TrimSpace(date)
	re := regexp.MustCompile(`[^0-9a-zA-Z']+`)
	parts := re.Split(udate, -1)
	if len(parts) < 3 {
		err = fmt.Errorf("not enough fields in date `%s`", date)
		return
	}
	parts = parts[:3]
	after = date
	for i, part := range parts {
		k := strings.Index(after, part)
		after = after[k+len(part):]
		parts[i] = strings.ToUpper(part)
	}

	year := ""
	month := ""
	day := ""

	for _, part := range parts {
		if year == "" && len(part) == 3 && strings.HasPrefix(part, "'") {
			part = "20" + part[1:]
		}
		if year == "" && len(part) == 4 && strings.TrimLeft(part, "1234567890") == "" {
			year = part
			continue
		}
		if year == "" && month == "" && day == "" && strings.TrimLeft(part, "1234567890") == "" {
			day = part
			continue
		}

		if year != "" && month != "" && day == "" && strings.TrimLeft(part, "1234567890") == "" {
			day = part
			continue
		}

		if (year != "" || day != "") && month == "" {
			month = part
			continue
		}
	}

	if year == "" || month == "" || day == "" {
		err = fmt.Errorf("`%s` is not a valid date [%s-%s-%s]", date, year, month, day)
		return
	}

	if len(month) > 2 {

		switch month[:3] {
		case "JAN":
			month = "1"
		case "FEB":
			month = "2"
		case "MAA":
			month = "3"
		case "MRT":
			month = "3"
		case "APR":
			month = "4"
		case "MEI":
			month = "5"
		case "JUN":
			month = "6"
		case "JUL":
			month = "7"
		case "AUG":
			month = "8"
		case "SEP":
			month = "9"
		case "OCT":
			month = "10"
		case "OKT":
			month = "10"
		case "NOV":
			month = "11"
		case "DEC":
			month = "12"
		}
	}

	if strings.TrimLeft(month, "1234567890") != "" {
		err = fmt.Errorf("`%s` is not a valid date", date)
		return
	}

	imonth, _ := strconv.ParseInt(month, 10, 0)
	if imonth < 1 || imonth > 12 {
		err = fmt.Errorf("`%s` has not a valid month(%d)", date, imonth)
		return
	}
	month = strings.TrimLeft(month, "0")
	iyear, _ := strconv.ParseInt(year, 10, 0)
	iday, _ := strconv.ParseInt(day, 10, 0)

	if iday < 1 {
		err = fmt.Errorf("`%s` has not a valid day(%d)", date, iday)
		return
	}

	if iday == 29 {
		if imonth == 2 && !IsLeap(iyear) {
			err = fmt.Errorf("`%s` has not a valid day(%d) for month %d", date, iday, imonth)
			return
		}
	}

	if iday > 31 {
		err = fmt.Errorf("`%s` has not a valid day(%d)", date, iday)
		return
	}

	if iday == 31 && strings.Index("1,3,5,7,8,10,12", month) < 0 {
		err = fmt.Errorf("`%s` has not 31 days in month %s", date, month)
		return
	}

	if iday == 30 && imonth == 2 {
		err = fmt.Errorf("`%s` has not 30 days in februari", date)
		return
	}
	tt := time.Date(int(iyear), time.Month(imonth), int(iday), 0, 0, 0, 0, time.Now().Location())
	t = &tt
	return
}

func IsLeap(year int64) bool {
	if year%4 != 0 {
		return false
	}
	if year%100 != 0 {
		return true
	}
	if year%400 == 0 {
		return true
	}
	return false
}
