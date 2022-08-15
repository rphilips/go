package manuscript

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"brocade.be/pbladng/chapter"
	pchapter "brocade.be/pbladng/chapter"
	ptools "brocade.be/pbladng/tools"
)

type Manuscript struct {
	Lines    []ptools.Line
	Start    int
	Year     int
	Week     int
	Bdate    *time.Time
	Edate    *time.Time
	Chapters []*chapter.Chapter
}

// New manuscript starting with a reader
func New(source io.Reader, pcode string) (m *Manuscript, err error) {
	m = new(Manuscript)
	blob, err := io.ReadAll(source)
	if err != nil {
		return nil, ptools.Error("manuscript-unreadable", 0, err.Error())
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
		year, week, bdate, edate, e := Header(s, lineno)
		if e != nil {
			return nil, e
		}
		m.Year = year
		m.Week = week
		m.Bdate = bdate
		m.Edate = edate
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
		if !strings.HasPrefix(s, "#") {
			return nil, ptools.Error("manuscript-prefix1", lineno, "line should start with `#`")
		}
		s = s[1:]
		s = strings.TrimSpace(s)
		if !strings.HasPrefix(s, "#") {
			return nil, ptools.Error("manuscript-prefix2", lineno, "line should start with `##`")
		}
		m.Start = lineno
		break
	}

	chapters, err := pchapter.Parse(m.Lines[m.Start:])

	if err != nil {

		m.Chapters = chapters
	}

	return m, err

}

func Header(line string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, err error) {
	line = strings.ToUpper(strings.TrimSpace(line))
	if !strings.HasPrefix(line, "WEEK") {
		err = ptools.Error("header-week", lineno, "first non-empty line should start with 'WEEK'")
		return
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
		err = ptools.Error("header-bdate", lineno, err.Error())
		return
	}
	after = strings.TrimLeft(after, "\t -")
	edate, after, err = ptools.NewDate(after)
	if err != nil {
		err = ptools.Error("header-edate", lineno, err.Error())
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
