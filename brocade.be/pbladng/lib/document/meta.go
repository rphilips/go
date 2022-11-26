package document

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	perror "brocade.be/pbladng/lib/error"
	pfs "brocade.be/pbladng/lib/fs"
	ptools "brocade.be/pbladng/lib/tools"
)

var reiso = regexp.MustCompile(`^20[0-9][0-9]-[0-9][0-9]$`)

func DocJSON(meta map[string]string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, mailed *time.Time, err error) {
	now := time.Now()

	if len(meta) == 0 {
		err = perror.Error("doc-meta-empty", lineno, "empty meta for document")
		return
	}
	value := meta["id"]
	y, w, ok := strings.Cut(value, "-")
	if !ok {
		err = perror.Error("doc-meta-id", lineno, "'week' should be of the form 'yyyy-ww'")
		return
	}
	year, e := strconv.Atoi(y)
	if e != nil {
		err = perror.Error("doc-meta-year", lineno, "'year' should be a number")
		return
	}
	week, e = strconv.Atoi(w)
	if e != nil {
		err = perror.Error("doc-meta-week", lineno, "'week' should be a number")
		return
	}

	for key, value := range meta {
		value := strings.TrimSpace(value)
		if value == "" && key != "mailed" {
			err = perror.Error("doc-meta-value-empty", lineno, "`"+key+"` is empty")
			return
		}
		switch key {
		case "id":
			_, _, ok := strings.Cut(value, "-")
			if !ok {
				err = perror.Error("doc-meta-id2", lineno, "'week' should be of the form 'yyyy-ww'")
				return
			}
			if year > (now.Year() + 1) {
				err = perror.Error("doc-meta-year2", lineno, "'year' should be smaller than next year")
				return
			}
			if year < 2022 {
				err = perror.Error("doc-meta-year3", lineno, "'year' should be greater than 2021")
				return
			}
			if week > 53 {
				err = perror.Error("doc-meta-weekmax", lineno, fmt.Sprintf("week %d should be smaller than 54", week))
				return
			}
			if week == 0 {
				err = perror.Error("doc-meta-weekmin", lineno, fmt.Sprintf("week %d should be not 0", week))
				return
			}
			tests := []string{y + "/" + w}
			switch {
			case week == 1:
				tests = append(tests, strconv.Itoa(year-1)+"/53", strconv.Itoa(year-1)+"/52")
			case week > 20 && week < 35:
				tests = append(tests, y+"/"+strconv.Itoa(week-1), y+"/"+strconv.Itoa(week-3))
			default:
				tests = append(tests, y+"/"+strconv.Itoa(week-1))
			}

			ok = false
			for _, f := range tests {
				f = "archive/manuscripts/" + f + "/week.pb"
				if pfs.Exists(f) {
					ok = true
					break
				}
			}
			if false && !ok {
				err = perror.Error("doc-meta-prevweek", lineno, fmt.Sprintf("week %d is invalid", week))
				return
			}
		case "bdate":
			bdate, _, err = ptools.NewDate(value)
			if err != nil {
				err = perror.Error("doc-meta-bdate", lineno, err)
				return
			}
			if week > 1 && bdate.Year() != year {
				err = perror.Error("doc-meta-bdate-year1", lineno, "year and bdate do not match")
				return
			}
			if week == 1 && (bdate.Year() != year && bdate.Year() != year-1) {
				err = perror.Error("doc-meta-bdate-year2", lineno, "year and bdate do not match")
				return
			}
		case "edate":
			edate, _, err = ptools.NewDate(value)
			if err != nil {
				err = perror.Error("doc-meta-edate", lineno, err)
				return
			}
			if week < 52 && edate.Year() != year {
				err = perror.Error("doc-meta-edate-year1", lineno, fmt.Sprintf("year %d and edate %d do not match ", year, edate.Year()))
				return
			}
			if week > 51 && (edate.Year() != year && edate.Year() != year+1) {
				err = perror.Error("doc-meta-edate-year2", lineno, "year and edate do not match")
				return
			}

		case "mailed":
			if value != "" {
				mailed, _, err = ptools.NewDate(value)
				if err != nil {
					err = perror.Error("doc-meta-mailed", lineno, err)
					return
				}
				if mailed.Year() < year {
					err = perror.Error("doc-meta-mailed-year1", lineno, "year and mailed do not match")
					return
				}
			}
		default:
			err = perror.Error("doc-meta-key", lineno, "`"+key+"` is unknown")
			return
		}
	}
	return

}

func TopicJSON(meta map[string]string, lineno int) (from *time.Time, until *time.Time, lastpb string, count int, maxcount int, notepb string, noteme string, ty string, err error) {

	if len(meta) == 0 {
		err = ptools.Error("topic-meta-empty", lineno, "empty meta for topic")
		return
	}
	for key, value := range meta {
		value := strings.TrimSpace(value)
		if value == "" {
			err = ptools.Error("topic-meta-value-empty", lineno, "`"+key+"` is empty")
			return
		}
		switch key {

		case "from":
			f, a, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("topic-meta-from-bad", lineno, e)
				return
			}
			if a != "" {
				err = ptools.Error("topic-meta-from-after", lineno, "trailing info after `from`")
				return
			}
			from = f

		case "lastpb":

			if !reiso.MatchString(value) {
				err = ptools.Error("topic-meta-lastpb-bad", lineno, "not a valid code")
				return
			}
			lastpb = value

		case "count":
			var e error
			count, e = strconv.Atoi(value)
			if e != nil {
				err = ptools.Error("topic-meta-count-bad", lineno, e)
				return
			}
		case "maxcount":
			var e error
			count, e = strconv.Atoi(value)
			if e != nil {
				err = ptools.Error("topic-meta-maxcount-bad", lineno, e)
				return
			}

		case "until":

			u, after, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("topic-meta-until-bad", lineno, e)
				return
			}
			if after != "" {
				err = ptools.Error("topic-meta-until-after", lineno, "trailing info after `until`")
				return
			}
			until = u

		case "notepb":
			notepb = value

		case "noteme":
			noteme = value
		case "type":
			value = strings.ToLower(value)
			if value != "cal" && value != "mass" {
				err = ptools.Error("topic-meta-type", lineno, "`"+value+"` is invalid type")
				return
			}
			ty = value

		default:
			err = ptools.Error("topic-meta-key", lineno, "`"+key+"` is unknown")
			return
		}
	}

	return

}