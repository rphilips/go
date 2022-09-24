package manuscript

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	pchapter "brocade.be/pbladng/lib/chapter"
	pfs "brocade.be/pbladng/lib/fs"
	pholy "brocade.be/pbladng/lib/holy"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
	ptopic "brocade.be/pbladng/lib/topic"
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

type ImageID struct {
	Mtime  string `json:"mtime"`
	Digest string `json:"digest"`
	Letter string `json:"letter"`
}

var rec = regexp.MustCompile(`^\s*[@#]\s*[@#]`)

func (m Manuscript) HTML(colofon bool, imgmanifest string) string {
	builder := strings.Builder{}
	esc := template.HTMLEscapeString
	dash := strings.Repeat("-", 96) + "<br />"
	props := pregistry.Registry["html"].(map[string]any)
	builder.WriteString("<!doctype html>\n<html lang='nl'>\n")
	builder.WriteString("<head>\n")
	builder.WriteString(fmt.Sprintf("<title>%s: week %s (%s - %s)</title>\n", esc(pregistry.Registry["pcode"].(string)), esc(m.ID()), esc(ptools.StringDate(m.Bdate, "I")), esc(ptools.StringDate(m.Edate, "I"))))
	builder.WriteString("</head>\n<body>\n")
	builder.WriteString(fmt.Sprintf("<b>Week: %s</b>", esc(m.ID())))
	builder.WriteString("<br />")
	builder.WriteString(fmt.Sprintf("<b>Editie: %s - %s</b>", esc(pregistry.Registry["pcode"].(string)), esc(props["title"].(string))))
	builder.WriteString("<br />")
	builder.WriteString(fmt.Sprintf("<b>%s</b><br />\n", strings.Repeat("-", 96)))

	if colofon {
		builder.WriteString(strings.Repeat("<br />", 3))
		builder.WriteString(dash)
		builder.WriteString("\n<b>OPGELET: NIEUWE COLOFON</b><br />")
		builder.WriteString(dash)
		builder.WriteString(strings.Repeat("<br />", 3))
		pcol := pfs.FName("support/colofon.txt")
		col, _ := bfs.Fetch(pcol)
		scol := strings.TrimSpace(string(col))
		builder.WriteString(scol)
		builder.WriteString("<br />")
	}

	bdate := m.Bdate
	edate := m.Edate

	imglist := make(map[string]ImageID)
	manifest, e := os.ReadFile(imgmanifest)
	imgletters := make(map[string]string)

	if e == nil {
		json.Unmarshal(manifest, &imglist)
		for fname, imgid := range imglist {
			imgletters[fname] = imgid.Letter
		}
	}

	for _, chapter := range m.Chapters {
		builder.WriteString(chapter.HTML(bdate, edate, m.ID(), imgletters))
	}
	builder.WriteString("</body></html>\n")

	return builder.String()
}

func (m Manuscript) String() string {
	builder := strings.Builder{}

	mailed := ""
	if m.Mailed != nil {
		mailed = ptools.StringDate(m.Mailed, "")
	}
	J := ptools.J

	h := fmt.Sprintf(`{ "id": %s, "bdate": %s, "edate": %s, "mailed": %s }`+"\n", J(m.ID()), J(ptools.StringDate(m.Bdate, "")), J(ptools.StringDate(m.Edate, "")), J(mailed))

	builder.WriteString(h)
	for _, chapter := range m.Chapters {
		builder.WriteString(chapter.String())
	}
	return builder.String()
}

func (m Manuscript) ID() string {
	return fmt.Sprintf("%d-%02d", m.Year, m.Week)
}

// New manuscript starting with a reader
func Parse(source io.Reader, checkextern bool, imgmanifest string) (m *Manuscript, err error) {
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
		c, err := pchapter.Parse(chap, m.ID(), m.Bdate, m.Edate, checkextern)
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

	if checkextern && imgmanifest != "" {
		letters := "abcdefghijklmnopqrstuvwxyz"
		imglist := make(map[string]ImageID)
		manifest, e := os.ReadFile(imgmanifest)

		if e == nil {
			json.Unmarshal(manifest, &imglist)
			for _, imgid := range imglist {
				letter := imgid.Letter
				if letter != "" {
					letters = strings.ReplaceAll(letters, letter, "")
				}
			}
		}
		counter := make(map[string]int)
		change := false
		for _, c := range m.Chapters {
			for _, t := range c.Topics {
				if len(t.Images) == 0 {
					continue
				}
				for _, img := range t.Images {
					fname := img.Fname
					if counter[fname] != 0 {
						err = ptools.Error("image-double1", img.Lineno, "same image as on line "+strconv.Itoa(counter[fname]))
						return
					}
					counter[fname] = img.Lineno
					digest := ""
					imgid := imglist[fname]
					mtime, digest, e := ptools.ImgProps(fname, img.Lineno, imgid.Mtime, imgid.Digest)
					if e != nil {
						delete(imglist, fname)
						bfs.Store(imgmanifest, imglist, "process")
						err = e
						return
					}
					for _, im := range t.Images {
						fn := im.Fname
						if fn == fname {
							continue
						}
						imgid, ok := imglist[fn]
						if !ok {
							continue
						}
						if imgid.Digest == digest {
							delete(imglist, fname)
							delete(imglist, fn)
							bfs.Store(imgmanifest, imglist, "process")
							err = ptools.Error("image-double2", img.Lineno, "same image as on line "+strconv.Itoa(im.Lineno))
							return
						}

					}
					change = true
					newletter := ""
					if letters != "" {
						newletter = letters[0:1]
						letters = letters[1:]
					}
					if newletter == "" {
						err = ptools.Error("image-above26", img.Lineno, "too many images")
						return
					}
					imglist[fname] = ImageID{
						Mtime:  mtime,
						Digest: digest,
						Letter: newletter,
					}
				}
			}
		}
		if change {
			bfs.Store(imgmanifest, imglist, "process")
		}

	}

	return
}

func Header(line string, lineno int) (year int, week int, bdate *time.Time, edate *time.Time, mailed *time.Time, err error) {
	line = strings.TrimSpace(line)
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
	line = strings.ToUpper(strings.TrimSpace(line))
	rem := regexp.MustCompile(`MAILED:\s*20\d\d-\d\d-\d\d`)
	mal := rem.FindString(line)
	if mal != "" {
		line = strings.ReplaceAll(line, mal, "")
		mal = strings.TrimSpace(strings.TrimPrefix(mal, "MAILED:"))
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
	matched, _ := regexp.MatchString(`^20\d\d-\d\d.*`, after)
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
	edate, _, err = ptools.NewDate(after)
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
		err = ptools.Error("header-json-invalid", lineno, "invalid JSON: "+e.Error())
		return
	}

	if len(mm) == 0 {
		err = ptools.Error("header-empty", lineno, "empty meta")
		return
	}
	value := mm["id"]
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

	week, e = strconv.Atoi(w)
	if e != nil {
		err = ptools.Error("header-week1-bad", lineno, "'week' should be a number")
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
			_, _, ok := strings.Cut(value, "-")
			if !ok {
				err = ptools.Error("header-week1-bad", lineno, "'week' should be of the form 'yyyy-ww'")
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
			if false && !ok {
				err = ptools.Error("header-prevweek", lineno, fmt.Sprintf("week %d is invalid", week))
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
				err = ptools.Error("header-edate-year1", lineno, fmt.Sprintf("year %d and edate %d do not match ", year, edate.Year()))
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

func Previous() (m *Manuscript) {
	_, m, err := FindBefore("", true)
	if err != nil {
		m = nil
	}
	return
}

func Next(m *Manuscript) (id string, year string, week string, bdate string, edate string) {
	if m == nil {
		return
	}
	startessential := pregistry.Registry["start-day-essential"].(string)
	startday := pregistry.Registry["start-day"].(string)

	date := m.Edate
	bd := m.Bdate

	for {
		if date.Before(*bd) {
			date = nil
			break
		}
		if date.Weekday().String() == startessential {
			break
		}
		d := date.AddDate(0, 0, -1)
		date = &d
	}
	if date == nil {
		return
	}

	bdn := date
	for {
		if bdn.Weekday().String() == startday {
			break
		}
		d := bdn.AddDate(0, 0, -1)
		bdn = &d
	}

	iyear := m.Year
	iweek := m.Week
	switch {
	case iweek < 52:
		iweek++
	case iweek > 52:
		iweek = 1
		iyear++
	case iweek == 52:
		if date.Year() == iyear {
			iweek = 53
		} else {
			iyear++
			iweek++
		}
	}

	ed := date

	inc, _ := strconv.Atoi(pregistry.Registry["last-day"].(string))
	for {
		if inc == 0 {
			break
		}
		inc--
		d := ed.AddDate(0, 0, 1)
		ed = &d
	}

	year = strconv.Itoa(iyear)
	week = strconv.Itoa(iweek)
	id = fmt.Sprintf("%d-%02d", iyear, iweek)
	bdate = ptools.StringDate(bdn, "I")

	edate = ptools.StringDate(ed, "I")
	return
}

var arcdir = pfs.FName("/archive/manuscripts")

func FindBefore(id string, mailed bool) (place string, m *Manuscript, err error) {
	if id == "" {
		now := time.Now()
		year := now.Year()
		id = strconv.Itoa(year)
	}
	if !strings.Contains(id, "-") {
		id = id + "-99"
	}
	syear, sweek, _ := strings.Cut(id, "-")
	year, err := strconv.Atoi(syear)
	if err != nil {
		return
	}
	// bfs.Store("/home/rphilips/Desktop/log.txt", id, "process")
	_, err = strconv.Atoi(sweek)
	if err != nil {
		return
	}

	for {
		if year < 2005 {
			return "", nil, fmt.Errorf("no manuscripts found")
		}
		dir := filepath.Join(arcdir, strconv.Itoa(year))

		files, err := os.ReadDir(dir)
		if err != nil {
			year--
			sweek = "99"
			continue
		}
		weeks := make([]string, 0)
		for _, w := range files {
			name := w.Name()
			base := filepath.Base(name)
			if len(name) != 2 {
				continue
			}
			if strings.TrimLeft(name, "1234567890") != "" {
				continue
			}
			if base < sweek {
				weeks = append(weeks, base)
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(weeks)))

		for _, week := range weeks {
			if week >= sweek {
				continue
			}
			fname := filepath.Join(dir, week, "week.pb")

			f, err := os.Open(fname)
			if err != nil {
				continue
			}
			source := bufio.NewReader(f)
			m, err := Parse(source, false, "")
			if err != nil {
				return "", nil, fmt.Errorf("error in %s: %s", fname, err.Error())
			}
			if !mailed || m.Mailed != nil {
				return fname, m, nil
			}
		}
		year--
		sweek = "99"
	}
}

func New(mm map[string]string, mold *Manuscript) (m *Manuscript, err error) {
	// {"action":"new","bdate":"2022-09-28","edate":"2022-10-02","week":"39","year":"2022"}
	m = new(Manuscript)
	//bfs.Store("/home/rphilips/Desktop/log.txt", mm, "process")

	m.Year, _ = strconv.Atoi(mm["year"])
	m.Week, _ = strconv.Atoi(mm["week"])
	m.Bdate, _, _ = ptools.NewDate(mm["bdate"])
	m.Edate, _, _ = ptools.NewDate(mm["edate"])
	m.Start = 1

	validti := pregistry.Registry["chapter-title-regexp"].([]any)
	for _, ti2 := range validti {
		ti := ti2.(map[string]any)["title"].(string)
		c := new(pchapter.Chapter)
		m.Chapters = append(m.Chapters, c)
		c.Header = ti
		ty := ti2.(map[string]any)["type"].(string)

		switch ty {
		case "new":
			c.Text = ti2.(map[string]any)["text"].(string)
		case "mass":
			t := new(ptopic.Topic)
			t.Header = ti2.(map[string]any)["header"].(string)
			t.Type = "mass"

			eudays := make([]*ptopic.Euday, 0)

			bdate := m.Bdate
			edate := m.Edate

			var told *ptopic.Topic
			for _, ch := range mold.Chapters {
				if ch.Header != ti {
					continue
				}
				if len(ch.Topics) == 0 {
					break
				}
				told = ch.Topics[0]
				if told.Type != "mass" {
					told = nil
					break
				}
				if len(told.Eudays) == 0 {
					told = nil
					break
				}
				break
			}
			last := bdate
			if told != nil {
				for _, euday := range told.Eudays {
					if euday.Date.Before(*bdate) {
						continue
					}
					if euday.Date.After(*edate) {
						break
					}
					last = euday.Date
					eudays = append(eudays, euday)
				}
			}
			date := last.AddDate(0, 0, -1)
			for {
				date = date.AddDate(0, 0, 1)
				if edate.Before(date) {
					break
				}
				weekday := date.Weekday().String()
				day := pregistry.Registry["mass-day"].(map[string]any)[weekday].([]any)
				if len(day) == 0 {
					continue
				}
				euday := new(ptopic.Euday)
				xdate := date
				euday.Date = &xdate
				euday.Headers = pholy.Today(&xdate)
				eudays = append(eudays, euday)
				for _, ms := range day {
					mms := ms.(map[string]any)
					st := mms["time"].(string)
					p := mms["place"].(string)
					i := mms["intention"].(string)
					if st == "" || p == "" || i == "" {
						continue
					}
					mass := new(ptopic.Mass)
					euday.M = append(euday.M, mass)
					if !strings.Contains(st, ".") {
						st += ".00"
					}
					sh, sm, _ := strings.Cut(st, ".")
					sh = strings.TrimLeft(sh, "0")
					sm = strings.TrimLeft(sm, "0")
					if sh == "" {
						sh = "0"
					}
					if sm == "" {
						sm = "0"
					}
					hour, _ := strconv.Atoi(sh)
					min, _ := strconv.Atoi(sm)

					ndate := time.Date(date.Year(), date.Month(), date.Day(), hour, min, 0, 0, date.Location())
					ndate.Hour()
					mass.Time = &ndate
					mass.Place = p
					for _, s := range strings.SplitN(i, "\n", -1) {
						s = strings.TrimSpace(s)
						if s == "" {
							continue
						}
						mass.Intentions = append(mass.Intentions, s)
					}
				}
			}
			t.Eudays = eudays
			c.Topics = append(c.Topics, t)

		default:
			for _, ch := range mold.Chapters {
				if ch.Header != ti {
					continue
				}
				for _, t := range ch.Topics {
					if t.MaxCount != 0 && t.MaxCount == t.Count {
						continue
					}
					if t.Until != nil && t.Until.Before(*m.Bdate) {
						continue
					}
					switch t.Type {
					case "cal":
						count := 0
						bdate := m.Bdate
						body := make([]ptools.Line, 0)
						oldbody := t.Body
						for _, line := range oldbody {
							s := line.L
							d := ptools.DetectDate(s)
							if d == nil {
								body = append(body, line)
								continue
							}
							if d.Before(*bdate) {
								continue
							}
							body = append(body, line)
							count++
						}
						if count != 0 {
							t.Body = body
							c.Topics = append(c.Topics, t)
						}
					default:
						c.Topics = append(c.Topics, t)
					}
				}
			}
		}

	}

	return
}
