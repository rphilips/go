package structure

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	bfs "brocade.be/base/fs"
	blines "brocade.be/base/lines"
	btime "brocade.be/base/time"
	perror "brocade.be/pbladng/lib/error"
	pfs "brocade.be/pbladng/lib/fs"
	pholy "brocade.be/pbladng/lib/holy"
	pnext "brocade.be/pbladng/lib/next"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

var chexp = regexp.MustCompile(`^[ \t]*#[ \t]*#.*$`)
var tpexp = regexp.MustCompile(`(?m)^[ \t]*#.*$`)

type Document struct {
	Lines    blines.Text
	Year     int
	Week     int
	Bdate    *time.Time
	Edate    *time.Time
	Mailed   *time.Time
	Colofon  bool
	Chapters []*Chapter
	Dir      string
	Letters  string
}

func (doc Document) HTML() string {

	builder := strings.Builder{}
	esc := template.HTMLEscapeString
	dash := strings.Repeat("-", 96) + "<br />"
	builder.WriteString("<!doctype html>\n<html lang='nl'>\n")
	builder.WriteString("<head>\n")
	builder.WriteString(fmt.Sprintf("<title>%s</title>\n", esc(doc.Title())))
	builder.WriteString("</head>\n<body>\n")
	builder.WriteString(fmt.Sprintf("<b>Week: %s</b>", esc(doc.ID())))
	builder.WriteString("<br />")
	builder.WriteString(fmt.Sprintf("<b>Editie: %s</b>", esc(doc.Title()[1:])))
	builder.WriteString("<br />")
	builder.WriteString(fmt.Sprintf("<b>%s</b><br />\n", strings.Repeat("-", 96)))

	if doc.Colofon {
		builder.WriteString(strings.Repeat("<br />", 3))
		builder.WriteString(dash)
		builder.WriteString("\n<b>OPGELET: NIEUW COLOFON</b><br />")
		builder.WriteString(dash)
		builder.WriteString(strings.Repeat("<br />", 3))
		pcol := pfs.FName("support/colofon.txt")
		col, err := bfs.Fetch(pcol)
		if err != nil {
			log.Fatalf("error in working with colofon at %s: %s", pcol, err)
		}
		scol := strings.TrimSpace(string(col))
		builder.WriteString(scol)
		builder.WriteString("<br />")
	}

	for _, chapter := range doc.Chapters {
		builder.WriteString(chapter.HTML())
	}
	builder.WriteString("</body></html>\n")

	return builder.String()

}
func (d Document) Title() string {
	edition := "first"
	if d.Mailed != nil {
		edition = "other"
	}
	ed := pregistry.Registry["edition"].(map[string]any)[edition].(map[string]any)
	subject := ed["subject"].(string)

	subject = strings.ReplaceAll(subject, "{week}", fmt.Sprintf("%02d", d.Week))
	subject = strings.ReplaceAll(subject, "{id}", d.ID())
	return subject
}

func (d Document) MailText() string {
	edition := "first"
	if d.Mailed != nil {
		edition = "other"
	}
	ed := pregistry.Registry["edition"].(map[string]any)[edition].(map[string]any)
	text := ed["body"].(string)

	text = strings.ReplaceAll(text, "{week}", fmt.Sprintf("%02d", d.Week))
	text = strings.ReplaceAll(text, "{id}", d.ID())
	return text
}

func (d Document) String() string {
	m := make(map[string]string)
	m["id"] = d.ID()
	m["bdate"] = btime.StringDate(d.Bdate, "I")
	m["edate"] = btime.StringDate(d.Edate, "I")
	if d.Colofon {
		m["colofon"] = "yes"
	} else {
		m["colofon"] = "no"
	}
	if d.Mailed != nil {
		m["mailed"] = btime.StringDate(d.Mailed, "I")
	} else {
		m["mailed"] = ""
	}
	j, _ := json.MarshalIndent(m, "", "    ")

	builder := strings.Builder{}

	builder.Write(j)
	for _, chapter := range d.Chapters {
		builder.WriteString(chapter.String())
	}
	return builder.String()
}

func (doc Document) LastChapter() (c *Chapter) {
	if len(doc.Chapters) == 0 {
		return
	}
	return doc.Chapters[len(doc.Chapters)-1]
}

func (doc Document) LastTopic() (t *Topic) {
	c := doc.LastChapter()
	if c == nil {
		return
	}
	return c.LastTopic()
}

func (doc Document) ArchiveDir() string {
	return pfs.FName(fmt.Sprintf("archive/manuscripts/%04d/%02d", doc.Year, doc.Week))
}

func (doc Document) ArchiveDirPrevious() string {
	fname, _, _ := FindBefore(doc.ID())
	if fname == "" {
		return ""
	}
	return filepath.Dir(fname)
}

func (doc Document) Archive() error {
	archive := doc.ArchiveDir()
	sourcedir := doc.Dir
	files, _, err := bfs.FilesDirs(sourcedir)
	if err != nil {
		return err
	}
	err = bfs.MkdirAll(archive, "process")
	if err != nil {
		return err
	}
	for _, f := range files {
		err = bfs.CopyFile(filepath.Join(sourcedir, f.Name()), archive, "", false)
		if err != nil {
			return err
		}
	}

	return err
}

func (doc Document) Names() (names []string) {
	for _, line := range doc.Lines {
		s := line.Text
		if !strings.ContainsRune(s, '|') {
			continue
		}
		rs := `\\`
		s = strings.ReplaceAll(s, rs, "")
		rs = `\|`
		s = strings.ReplaceAll(s, rs, "")
		if !strings.ContainsRune(s, '|') {
			continue
		}
		parts := strings.SplitN(s, "|", -1)
		for i, part := range parts {
			part, _ := ptools.Normalize(part, true)
			if part == "" {
				continue
			}
			if i%2 == 0 {
				continue
			}
			names = append(names, part)
		}
	}
	return
}
func (doc Document) Next() (id string, year string, week string, bdate *time.Time, edate *time.Time) {

	id, date := pnext.NextToNew(doc.ID())
	if id == "" {
		return
	}
	year, week, _ = strings.Cut(id, "-")
	bdate = btime.DetectDate(date)
	startday := pregistry.Registry["start-day"].(string)
	for {
		if bdate.Weekday().String() == startday {
			break
		}
		x := bdate.AddDate(0, 0, 1)
		bdate = &x
	}
	id2, date := pnext.NextToNew(id)
	if id2 == "" {
		return
	}
	edate = btime.DetectDate(date)
	lastday := pregistry.Registry["last-day"].(string)
	for {
		if edate.Weekday().String() == lastday {
			break
		}
		x := edate.AddDate(0, 0, -1)
		edate = &x
	}
	return
}

func FindBefore(id string) (fname string, doc *Document, err error) {
	if id == "" {
		now := time.Now()
		year := now.Year()
		_, week := now.ISOWeek()
		week += 2
		id = fmt.Sprintf("%04d-%02d", year, week)
	}
	if !strings.Contains(id, "-") {
		id = id + "-54"
	}
	syear, sweek, _ := strings.Cut(id, "-")
	year, err := strconv.Atoi(syear)
	if err != nil {
		return
	}
	week, err := strconv.Atoi(sweek)
	if err != nil {
		return
	}

	if week == 1 {
		week = 54
		year = year - 1
	}
	fname = pfs.FName(fmt.Sprintf("archive/manuscripts/%04d/%02d/week.pb", year, week-1))
	if !bfs.IsFile(fname) {
		fname = ""
	}
	if fname == "" && !bfs.IsDir(pfs.FName(fmt.Sprintf("archive/manuscripts/%04d", year))) {
		week = 54
		year = year - 1
		if !bfs.IsDir(pfs.FName(fmt.Sprintf("archive/manuscripts/%04d", year))) {
			err = fmt.Errorf("cannot find previous pblad")
			return "", nil, err
		}
	}
	if fname == "" {
		for i := week - 1; i > 0; i-- {
			try := pfs.FName(fmt.Sprintf("archive/manuscripts/%04d/%02d/week.pb", year, i))
			if bfs.IsFile(try) {
				fname = try
				break
			}
		}
	}
	if fname == "" {
		err = fmt.Errorf("cannot find previous pblad")
		return "", nil, err
	}
	fmt.Println(fname)
	f, err := os.Open(fname)
	if err != nil {
		return "", nil, err
	}
	source := bufio.NewReader(f)

	doc = new(Document)
	doc.Dir = filepath.Dir(fname)
	err = doc.Load(source)
	if err != nil {
		return "", nil, fmt.Errorf("cannot load document at %s", fname)
	}
	return fname, doc, nil
}

func (doc Document) ID() string {
	return fmt.Sprintf("%04d-%02d", doc.Year, doc.Week)
}

func (doc *Document) Load(source io.Reader) error {
	err := doc.LoadText(source, true)
	if err != nil {
		return err
	}
	err = doc.LoadMeta()
	if err != nil {
		return err
	}

	err = doc.LoadChapters()

	return err
}

func (doc *Document) LoadText(source io.Reader, latin1 bool) error {
	buf := bufio.NewReader(source)
	cr := byte('\n')
	repl := rune(65533)
	lineno := 0
	t := make([]blines.Line, 0)

	for {
		lineno++
		b, rerr := buf.ReadBytes(cr)
		if rerr != nil && rerr != io.EOF {
			err := perror.Error("docload-read", lineno, rerr)
			return err
		}
		if !utf8.Valid(b) {
			err := perror.Error("docload-noutf8", lineno, "line is not valid UTF-8")
			return err
		}
		if bytes.ContainsRune(b, repl) {
			err := perror.Error("docload-repl", lineno, "unicode replacement character in line")
			return err
		}
		line := blines.Line{}
		line.Lineno = lineno
		if latin1 {
			line.Text = ptools.NormalizeR(string(b), true)
		} else {
			line.Text = string(b)
		}
		t = append(t, line)
		if rerr == io.EOF {
			break
		}
	}
	doc.Lines = blines.Compact(t)
	return nil
}

func (doc *Document) LoadMeta() error {
	if len(doc.Lines) == 0 {
		err := perror.Error("docmeta-empty", 0, "no data available")
		return err
	}
	first := doc.Lines[0].Text
	lineno1 := doc.Lines[0].Lineno
	if !strings.HasPrefix(first, "{") {
		err := perror.Error("docmeta-open", lineno1, "line should start with `{`")
		return err
	}
	end := -1
	slicemeta := make([]string, 0, len(doc.Lines))
	for i, line := range doc.Lines {
		slicemeta = append(slicemeta, line.Text)
		if strings.ContainsRune(line.Text, '}') {
			end = i
			break
		}
	}
	if end == -1 {
		err := perror.Error("docmeta-close", 0, "data does not contain `}`")
		return err
	}
	last := doc.Lines[end].Text
	lineno2 := doc.Lines[end].Lineno
	if !strings.HasSuffix(last, "}") {
		err := perror.Error("docmeta-close", lineno2, "line should end with `}`")
		return err
	}
	smeta := strings.Join(slicemeta, "\n")
	meta := make(map[string]string)
	err := json.Unmarshal([]byte(smeta), &meta)
	if err != nil {
		err := perror.Error("docmeta-unmarshal", lineno1, "no valid json")
		return err
	}

	findLineno := func(key string) int {
		rexp := regexp.MustCompile(`"\s*` + regexp.QuoteMeta(key) + `\s*"\s*:`)
		index := blines.Index(doc.Lines, rexp, lineno1, lineno2)
		if index < 0 {
			return -1
		}
		return doc.Lines[index].Lineno
	}

	for key, value := range meta {
		value := strings.TrimSpace(value)
		key := strings.TrimSpace(key)
		if key != "id" {
			continue
		}
		now := time.Now()
		syear, sweek, ok := strings.Cut(value, "-")
		if !ok {
			lineno := findLineno(key)
			err := perror.Error("docmeta-id", lineno, "'id' should be of the form 'yyyy-ww'")
			return err
		}
		year, err := strconv.Atoi(syear)
		if err != nil {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idyear-form", lineno, "'year' should be of the form 'yyyy'")
			return err
		}
		if year > (now.Year() + 1) {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idyear-big", lineno, "'year' should be smaller than next year")
			return err
		}
		if year < 2023 {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idyear-bad", lineno, "'year' should be greater than 2022")
			return err
		}
		week, err := strconv.Atoi(sweek)
		if err != nil {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idweek-form", lineno, "'year' should be of the form 'ww'")
			return err
		}
		if week > 53 {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idweek-max", lineno, fmt.Sprintf("week %d should be smaller than 54", week))
			return err
		}
		if week < 1 {
			lineno := findLineno(key)
			err := perror.Error("docmeta-idweek-min", lineno, fmt.Sprintf("week %d should not be less than 1", week))
			return err
		}
		id := fmt.Sprintf("%04d-%02d", year, week)
		send, _ := pnext.Special(id)
		if send == nil {
			lineno := findLineno(key)
			err := perror.Error("docmeta-id-send", lineno, fmt.Sprintf("id %s not scheduled", id))
			return err
		}
		doc.Year = year
		doc.Week = week
		break
	}
	if doc.Year == 0 || doc.Week == 0 {
		lineno := findLineno("id")
		err := perror.Error("docmeta-id-notfound", lineno, "id is missing in document meta")
		return err
	}
	for key, value := range meta {
		value := strings.TrimSpace(value)
		key := strings.TrimSpace(key)
		if key == "id" {
			continue
		}

		if value == "" && key != "mailed" {
			lineno := findLineno(key)
			err := perror.Error("docmeta-emptyvalue", lineno, "`"+key+"` is empty")
			return err
		}
		switch key {
		case "bdate":
			bdate := btime.DetectDate(value)
			if bdate == nil {
				lineno := findLineno(key)
				err := perror.Error("docmeta-bdate-date", lineno, "invalid date")
				return err
			}
			if doc.Week > 1 && bdate.Year() != doc.Year {
				lineno := findLineno(key)
				err := perror.Error("docmeta-bdate-year1", lineno, "year and bdate do not match")
				return err
			}
			if doc.Week == 1 && (bdate.Year() != doc.Year && bdate.Year() != doc.Year-1) {
				lineno := findLineno(key)
				err := perror.Error("docmeta-bdate-year2", lineno, "year and bdate do not match")
				return err
			}
			doc.Bdate = bdate
		case "edate":
			edate := btime.DetectDate(value)
			if edate == nil {
				lineno := findLineno(key)
				err := perror.Error("docmeta-edate", lineno, "invalid date")
				return err
			}
			if doc.Week < 52 && edate.Year() != doc.Year {
				lineno := findLineno(key)
				err := perror.Error("docmeta-edate-year1", lineno, fmt.Sprintf("year %d and edate %d do not match ", doc.Year, edate.Year()))
				return err
			}
			if doc.Week > 51 && (edate.Year() != doc.Year && edate.Year() != doc.Year+1) {
				lineno := findLineno(key)
				err := perror.Error("heading-edate-year2", lineno, "year and edate do not match")
				return err
			}
			doc.Edate = edate
		case "colofon":
			doc.Colofon = ptools.IsTrue(value)

		case "mailed":
			if value != "" {
				mailed := btime.DetectDate(value)
				if mailed == nil {
					lineno := findLineno(key)
					err := perror.Error("docmeta-mailed", lineno, "invalid date")
					return err
				}
				if mailed.Year() < doc.Year {
					lineno := findLineno(key)
					err := perror.Error("docmeta-mailed-year1", lineno, "year and mailed do not match")
					return err
				}
				doc.Mailed = mailed
			}
		default:
			lineno := findLineno(key)
			err = perror.Error("docmeta-key", lineno, "`"+key+"` is unknown")
			return err
		}
	}
	if doc.Bdate == nil {
		lineno := findLineno("bdate")
		if lineno < 1 {
			lineno = 0
		}
		err := perror.Error("docmeta-bdate-missing", lineno, "bdate is missing")
		return err
	}

	if doc.Edate == nil {
		lineno := findLineno("edate")
		if lineno < 1 {
			lineno = 0
		}
		err := perror.Error("docmeta-edate-missing", lineno, "edate is missing")
		return err
	}

	if doc.Edate.Before(*doc.Bdate) {
		lineno := findLineno("edate")
		if lineno < 1 {
			lineno = 0
		}
		err := perror.Error("docmeta-edate-late", lineno, "edate is before bdate")
		return err
	}

	return nil

}

func (doc *Document) LoadChapters() error {
	ts := blines.Split(doc.Lines, chexp)
	if len(ts) < 2 {
		err := perror.Error("docchapter-notfound", 0, "no chapters found")
		return err
	}
	first := blines.Compact(ts[0])

	if len(first) == 0 {
		err := perror.Error("docchapter-nometa", 0, "no meta found")
		return err
	}
	last := first[len(first)-1]
	if !strings.HasSuffix(last.Text, "}") {
		err := perror.Error("docchapter-notmeta", last.Lineno, "should end on `}`")
		return err
	}

	for i := 1; i < len(ts); i += 2 {
		tc := append(ts[i], ts[i+1]...)
		c := new(Chapter)
		c.Document = doc
		doc.Chapters = append(doc.Chapters, c)
		err := c.Load(tc)
		if err != nil {
			return err
		}
	}
	sort.Slice(doc.Chapters, func(i, j int) bool { return doc.Chapters[i].Sort < doc.Chapters[j].Sort })

	prevheading := ""
	prevlineno := 0
	for _, c := range doc.Chapters {
		if c.Heading == prevheading {
			err := perror.Error("chapter-title-double", c.Lineno, fmt.Sprintf("chapter title also found at line %d", prevlineno))
			return err
		}
		prevheading = c.Heading
		prevlineno = c.Lineno
	}

	return nil

}

func New(mm map[string]string, dold *Document) (doc *Document, err error) {
	// {"action":"new","bdate":"2022-09-28","edate":"2022-10-02","week":"39","year":"2022"}
	doc = new(Document)
	//bfs.Store("/home/rphilips/Desktop/log.txt", mm, "process")

	doc.Year, _ = strconv.Atoi(mm["year"])
	doc.Week, _ = strconv.Atoi(mm["week"])
	doc.Bdate = btime.DetectDate(mm["bdate"])
	if strings.HasPrefix(mm["edate"], "3") {
		mm["edate"] = "2" + mm["edate"][1:]
	}
	doc.Edate = btime.DetectDate(mm["edate"])

	validti := pregistry.Registry["chapter-title-regexp"].([]any)
	for _, ti2 := range validti {
		ti := ti2.(map[string]any)["heading"].(string)
		c := new(Chapter)
		c.Document = doc
		doc.Chapters = append(doc.Chapters, c)
		c.Heading = ti
		ty := ti2.(map[string]any)["type"].(string)
		switch ty {
		case "new":
			body := ti2.(map[string]any)["text"].(string)
			tx := blines.ConvertString(body, 1)
			tx = blines.Compact(tx)
			c.Load(tx)
		case "mass":
			t := new(Topic)
			t.Heading = ti2.(map[string]any)["heading"].(string)
			t.Type = "mass"
			t.Chapter = c
			c.Topics = append(c.Topics, t)
			eudays := make([]*Euday, 0)

			bdate := doc.Bdate
			edate := doc.Edate

			var told *Topic
			for _, ch := range dold.Chapters {
				if len(ch.Topics) == 0 {
					told = nil
					continue
				}
				told = ch.Topics[0]
				if told.Type != "mass" {
					told = nil
					continue
				}
				if len(told.Eudays) == 0 {
					told = nil
					continue
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
				euday := new(Euday)
				xdate := date
				euday.Date = &xdate
				euday.Headings = pholy.Today(&xdate)
				eudays = append(eudays, euday)
				for _, ms := range day {
					mms := ms.(map[string]any)
					st := mms["time"].(string)
					p := mms["place"].(string)
					i := mms["intention"].(string)
					if st == "" || p == "" || i == "" {
						continue
					}
					mass := new(Mass)
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

		default:
			for _, ch := range dold.Chapters {
				if ch.Heading != ti {
					continue
				}
				for _, t := range ch.Topics {
					if t.MaxCount != 0 && t.MaxCount == t.Count {
						continue
					}
					if t.Until != nil && t.Until.Before(*doc.Bdate) {
						continue
					}
					switch t.Type {
					case "cal":
						count := 0
						bdate := doc.Bdate
						body := make([]blines.Line, 0)
						oldbody := t.Body
						for _, line := range oldbody {
							s := line.Text
							if strings.HasPrefix(s, "//") {
								body = append(body, line)
								count++
								continue
							}
							d := btime.DetectDate(s)
							if d == nil {
								body = append(body, line)
								count++
								continue
							}
							if d.Before(*bdate) {
								continue
							}
							body = append(body, line)
							count++
						}
						if count != 0 {
							t.Body = blines.Compact(body)
							t.Chapter = c
							c.Topics = append(c.Topics, t)
						}
					default:
						t.Chapter = c
						c.Topics = append(c.Topics, t)
					}
				}
			}
		}

	}

	return
}
