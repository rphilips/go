package manuscript

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	pchapter "brocade.be/pbladng/lib/chapter"
	pfs "brocade.be/pbladng/lib/fs"
	ptools "brocade.be/pbladng/lib/tools"
)

type Manuscript struct {
	Lines    []ptools.Line
	Start    int
	Year     int
	Week     int
	Bdate    *time.Time
	Edate    *time.Time
	Mailed   *time.Time
	Chapters []*pchapter.Chapter
}

var rec = regexp.MustCompile(`^\s*[@#]\s*[@#]`)

func (m Manuscript) String() string {
	builder := strings.Builder{}

	mailed := ""
	if m.Mailed != nil {
		mailed = ptools.StringDate(m.Mailed, "")
	}
	mm := map[string]string{
		"id":     m.ID(),
		"bdate":  ptools.StringDate(m.Bdate, ""),
		"edate":  ptools.StringDate(m.Edate, ""),
		"mailed": mailed,
	}
	h, _ := json.Marshal(mm)
	builder.Write(h)
	builder.WriteString("\n")
	for _, chapter := range m.Chapters {
		builder.WriteString(chapter.String())
	}
	return builder.String()
}

func (m Manuscript) ID() string {
	return fmt.Sprintf("%d-%02d", m.Year, m.Week)
}

// New manuscript starting with a reader
func Parse(source io.Reader) (m *Manuscript, err error) {
	m = new(Manuscript)
	blob, err := io.ReadAll(source)
	if err != nil {
		return nil, ptools.Error("manuscript-unreadable", 0, err)
	}
	lines, err := ptools.IsUTF8(blob)
	if err != nil {
		return
	}
	m.Lines = make([]ptools.Line, len(lines))
	for i, line := range lines {
		m.Lines[i] = ptools.Line{
			L:  line,
			NR: i + 1,
		}
	}

	start := 0
	for _, line := range m.Lines {
		lineno := line.NR
		s := strings.TrimSpace(line.L)
		// WEEK 2022-32 (2022-08-13 - 2022-08-21; lectors: 0; jgts: 0)
		if s == "" {
			continue
		}
		year, week, bdate, edate, mailed, e := Header(s, lineno)
		if e != nil {
			return nil, e
		}
		m.Year = year
		m.Week = week
		m.Bdate = bdate
		m.Edate = edate
		m.Mailed = mailed
		start = lineno
		break
	}

	for _, line := range m.Lines {
		lineno := line.NR
		if lineno <= start {
			continue
		}
		s := strings.TrimSpace(line.L)
		if s == "" {
			continue
		}
		if !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, "@") {
			return nil, ptools.Error("manuscript-fluff1", lineno, "line should start with `##`")
		}
		s = s[1:]
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, "@") {
			return nil, ptools.Error("manuscript-fluff2", lineno, "line should start with `##`")
		}
		m.Start = lineno
		break
	}

	if m.Start == 0 {
		return nil, ptools.Error("manuscript-empty", 0, "manuscript is empty")
	}

	chaps := make([][]ptools.Line, 0)
	for _, line := range m.Lines[m.Start-1:] {
		s := strings.TrimSpace(line.L)
		if s == "" && (len(chaps) == 0 || len(chaps[len(chaps)-1]) == 0) {
			continue
		}
		nieuw := rec.MatchString(s)
		if s != "" && len(chaps) == 0 && !nieuw {
			err = ptools.Error("chapter-fluff", line.NR, "text outside of chapter")
			return
		}
		if nieuw {
			chaps = append(chaps, make([]ptools.Line, 0))
		}
		chaps[len(chaps)-1] = append(chaps[len(chaps)-1], line)
	}

	if len(chaps) == 0 {
		return
	}

	for _, chap := range chaps {
		c, err := pchapter.Parse(chap, m.ID(), m.Bdate, m.Edate)
		if err != nil {
			return nil, err
		}
		m.Chapters = append(m.Chapters, c)
	}

	doubles := make(map[string]int)
	for _, c := range m.Chapters {
		ti := ptools.HeaderString(c.Header)
		k := strings.Index(ti, "[")
		if k != -1 {
			ti = strings.TrimSpace(ti[:k])
		}
		if doubles[ti] != 0 {
			err = ptools.Error("chapter-double", c.Start, "title occured on line "+strconv.Itoa(doubles[ti]))
			return
		}
		doubles[ti] = c.Start
	}

	if len(m.Chapters) != 0 {
		sort.Slice(m.Chapters, func(i, j int) bool { return m.Chapters[i].Sort < m.Chapters[j].Sort })
	}
	return
}

func Header(line string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, mailed *time.Time, err error) {
	line = strings.ToUpper(strings.TrimSpace(line))
	switch {
	case strings.HasPrefix(line, "{"):
		return HeaderJ(line, lineno)
	case strings.HasPrefix(line, "WEEK"):
		return HeaderT(line, lineno)
	default:
		err = ptools.Error("header-week", lineno, "first non-empty line should start with 'WEEK' or with '{'")
		return
	}
}

func HeaderT(line string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, mailed *time.Time, err error) {
	rem := regexp.MustCompile(`mailed:\s*20\d\d-\d\d-\d\d`)
	mal := rem.FindString(line)
	if mal != "" {
		line = strings.ReplaceAll(line, mal, "")
		mal = strings.TrimSpace(strings.TrimPrefix(line, "mailed:"))
		x, _, e := ptools.NewDate(mal)
		if e == nil {
			mailed = x
		}
	}
	after := strings.TrimPrefix(line, "WEEK")
	after = strings.TrimSpace(after)
	if after == "" {
		err = ptools.Error("header-after", lineno, "first non-empty line should contain more information")
		return
	}
	matched, err := regexp.MatchString(`^20\d\d-\d\d.*`, after)
	if !matched {
		err = ptools.Error("header-yearweek", lineno, "first non-empty line should start with 'WEEK yyyy-ww")
		return
	}
	year, _ = strconv.Atoi(after[:4])
	week, _ = strconv.Atoi(after[5:7])

	if week > 53 {
		err = ptools.Error("header-weekmax", lineno, fmt.Sprintf("week %d should be smaller than 54", week))
		return
	}
	after = after[7:]
	after = strings.TrimLeft(after, "\t :(")
	bdate, after, err = ptools.NewDate(after)
	if err != nil {
		err = ptools.Error("header-bdate", lineno, err)
		return
	}
	after = strings.TrimLeft(after, "\t -")
	edate, after, err = ptools.NewDate(after)
	if err != nil {
		err = ptools.Error("header-edate", lineno, err)
		return
	}
	if week > 1 && bdate.Year() != year {
		err = ptools.Error("header-bdate-year1", lineno, "year and bdate do not match")
		return
	}
	if week == 1 && (bdate.Year() != year && bdate.Year() != year-1) {
		err = ptools.Error("header-bdate-year2", lineno, "year and bdate do not match")
		return
	}
	if week < 52 && edate.Year() != year {
		err = ptools.Error("header-edate-year1", lineno, "year and edate do not match")
		return
	}
	if week > 51 && (edate.Year() != year && edate.Year() != year+1) {
		err = ptools.Error("header-edate-year2", lineno, "year and edate do not match")
		return
	}
	return

}

func HeaderJ(line string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, mailed *time.Time, err error) {
	now := time.Now()
	line = strings.TrimSpace(line)
	mm := make(map[string]string)
	e := json.Unmarshal([]byte(line), &mm)
	if e != nil {
		err = ptools.Error("header-json-invalid", lineno, "invalid JSON")
		return
	}

	if len(mm) == 0 {
		err = ptools.Error("header-empty", lineno, "empty meta")
		return
	}
	for key, value := range mm {
		value := strings.TrimSpace(value)
		if value == "" && key != "mailed" {
			err = ptools.Error("header-value-empty", lineno, "`"+key+"` is empty")
			return
		}
		switch key {
		case "id":
			y, w, ok := strings.Cut(value, "-")
			if !ok {
				err = ptools.Error("header-week1-bad", lineno, "'week' should be of the form 'yyyy-ww'")
				return
			}
			year, e = strconv.Atoi(y)
			if e != nil {
				err = ptools.Error("header-year1-bad", lineno, "'year' should be a number")
				return
			}
			if year > (now.Year() + 1) {
				err = ptools.Error("header-year2-bad", lineno, "'year' should be smaller than next year")
				return
			}
			if year < 2022 {
				err = ptools.Error("header-year3-bad", lineno, "'year' should be greater than 2021")
				return
			}
			week, e = strconv.Atoi(w)
			if e != nil {
				err = ptools.Error("header-week1-bad", lineno, "'week' should be a number")
				return
			}

			if week > 53 {
				err = ptools.Error("header-weekmax", lineno, fmt.Sprintf("week %d should be smaller than 54", week))
				return
			}
			if week == 0 {
				err = ptools.Error("header-weekmin", lineno, fmt.Sprintf("week %d should be not 0", week))
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
			if !ok {
				err = ptools.Error("header-week", lineno, fmt.Sprintf("week %d is invalid", week))
				return
			}
		case "bdate":
			bdate, _, err = ptools.NewDate(value)
			if err != nil {
				err = ptools.Error("header-bdate", lineno, err)
				return
			}
			if week > 1 && bdate.Year() != year {
				err = ptools.Error("header-bdate-year1", lineno, "year and bdate do not match")
				return
			}
			if week == 1 && (bdate.Year() != year && bdate.Year() != year-1) {
				err = ptools.Error("header-bdate-year2", lineno, "year and bdate do not match")
				return
			}
		case "edate":
			edate, _, err = ptools.NewDate(value)
			if err != nil {
				err = ptools.Error("header-edate", lineno, err)
				return
			}
			if week < 52 && edate.Year() != year {
				err = ptools.Error("header-edate-year1", lineno, "year and edate do not match")
				return
			}
			if week > 51 && (edate.Year() != year && edate.Year() != year+1) {
				err = ptools.Error("header-edate-year2", lineno, "year and edate do not match")
				return
			}

		case "mailed":
			if value != "" {
				mailed, _, err = ptools.NewDate(value)
				if err != nil {
					err = ptools.Error("header-mailed", lineno, err)
					return
				}
				if mailed.Year() < year {
					err = ptools.Error("header-mailed-year1", lineno, "year and mailed do not match")
					return
				}
			}
		default:
			err = ptools.Error("header-key", lineno, "`"+key+"` is unknown")
			return
		}
	}
	return

}

func Previous() (id string, period string, mailed string) {
	return "1999-10", "2022-01-15 - 2022-01-25", "2022-01-12"
}

func Next() (year string, week string, bdate string, edate string) {
	return "2000", "13", "2000-02-15", "2000-02-25"
}
