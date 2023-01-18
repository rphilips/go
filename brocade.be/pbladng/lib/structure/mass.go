package structure

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	blines "brocade.be/base/lines"
	bstrings "brocade.be/base/strings"
	btime "brocade.be/base/time"
	perror "brocade.be/pbladng/lib/error"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

type Euday struct {
	Date     *time.Time
	Headings []string
	Start    int
	M        []*Mass
}

type Mass struct {
	Time       *time.Time
	Place      string
	Lectors    []string
	Dealers    []string
	Intentions []string
}

func (m Mass) String() string {
	hour := m.Time.Hour()
	min := m.Time.Minute()
	place, _ := findPlace(m.Place)
	lecs := make([]string, len(m.Lectors))
	for i, lec := range m.Lectors {
		_, lecs[i] = findPerson(place, lec)
	}
	dels := make([]string, len(m.Dealers))
	for i, del := range m.Dealers {
		_, dels[i] = findPerson(place, del)
	}
	lector := strings.Join(lecs, ";")
	dealer := strings.Join(dels, ";")

	if dealer != "" {
		lector += " / " + dealer
	}

	lector = strings.TrimSpace(lector)
	if lector != "" {
		lector = " " + lector
	}
	ints := strings.Join(m.Intentions, "\n")
	ints = strings.TrimSpace(ints)
	if ints != "" {
		ints = " " + ints
	}
	return fmt.Sprintf("\n%02d.%02d %s%s:%s", hour, min, place, lector, ints)
}

func (m Mass) HTML() string {
	hour := m.Time.Hour()
	min := m.Time.Minute()
	_, place := findPlace(m.Place)
	lecs := make([]string, len(m.Lectors))
	for i, lec := range m.Lectors {
		_, lecs[i] = findPerson(m.Place, lec)
	}
	dels := make([]string, len(m.Dealers))
	for i, del := range m.Dealers {
		_, dels[i] = findPerson(m.Place, del)
	}
	lector := strings.Join(lecs, ", ")
	dealer := strings.Join(dels, ", ")

	ints := strings.Join(m.Intentions, "\n")
	ints = strings.TrimSpace(ints)
	uints := strings.TrimSpace(strings.ReplaceAll("\n"+strings.ToUpper(ints)+"\n", "EUCHARISTIE", ""))
	if uints != "" {
		euch := regexp.MustCompile("\n\\s*[Ee][Uu][Cc][Hh][Aa][Rr][Ii][Ss][Tt][Ii][Ee]\\s*\n")
		ints = euch.ReplaceAllString("\n"+ints+"\n", "\n")
		ints = strings.ReplaceAll(ints, "\n\n", "\n")
	}
	ints = strings.TrimSpace(ints)

	if ints != "" {
		ints = " " + ints
	}
	esc := template.HTMLEscapeString
	h := ptools.Html
	ints = h(esc(ints)) + "\n"
	s := fmt.Sprintf("\n<i>%02d.%02d u. %s</i>:%s", hour, min, place, ints)
	if lector != "" {
		s += "Lector: " + h(esc(lector)) + "\n"
	}
	if dealer != "" {
		s += "Communiedeler: " + h(esc(dealer)) + "\n"
	}
	s = strings.TrimSpace(s)
	return strings.ReplaceAll(s, "\n", "<br />")
}

func (d Euday) String() string {
	day := d.Date.Day()
	month := d.Date.Month()
	weekday, _, _ := strings.Cut(btime.StringDate(d.Date, "D"), " ")
	weekday = strings.ToUpper(weekday[0:1]) + weekday[1:]

	headings := strings.Join(d.Headings, "\n")
	headings = strings.TrimSpace(headings)

	if headings != "" {
		headings = headings + "\n"
	}

	mas := ""
	for _, m := range d.M {
		mas += strings.TrimSpace(m.String()) + "\n"
	}
	return strings.TrimSpace(fmt.Sprintf("%s %02d/%02d\n%s%s", weekday, day, month, headings, mas))
}

func (d Euday) HTML() string {
	weekday := btime.StringDate(d.Date, "D")
	weekday = strings.ToUpper(weekday[0:1]) + weekday[1:]
	esc := template.HTMLEscapeString
	h := ptools.Html
	headings := strings.Join(d.Headings, "\n")
	headings = strings.TrimSpace(headings)
	if headings != "" {
		headings = h(esc(headings))
		headings = strings.ReplaceAll(headings, "\n", "<br />")
		headings += "<br />"
	}
	mas := ""
	for _, m := range d.M {
		mas += m.HTML() + "<br />"
	}
	return strings.TrimSpace(fmt.Sprintf("<b>%s</b><br />%s%s", weekday, headings, mas))
}

func (t *Topic) LoadMass() error {
	if t.Type != "mass" {
		return nil
	}
	doc := t.Chapter.Document
	bdate := doc.Bdate
	edate := doc.Edate

	dys := make([]blines.Text, 0)
	days := make([]*Euday, 0)

	rem1 := regexp.MustCompile(`^[A-Z][a-z]+dag [0-9]+ [a-z]+ [0-9]{4}$`)
	rem2 := regexp.MustCompile(`^[A-Z][a-z]+dag [0-9]{1,2}/[0-9]{1,2}$`)

	for _, line := range t.Body {
		s := strings.ReplaceAll(line.Text, "*", "")
		s, _, _ = strings.Cut(s, "[")
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		nieuw := rem1.MatchString(s) || rem2.MatchString(s)
		if !nieuw && len(dys) == 0 {
			err := perror.Error("day-fluff", line.Lineno, "text outside of day")
			return err
		}
		if nieuw {
			dys = append(dys, make(blines.Text, 0))
		}
		dys[len(dys)-1] = append(dys[len(dys)-1], line)
	}
	if len(dys) == 0 {
		err := perror.Error("day-missing", t.Lineno, "mass missing in day")
		return err
	}
	// per dag
	for _, dy := range dys {
		if len(dy) < 2 {
			err := perror.Error("day-missing", dy[0].Lineno, "info missing")
			return err
		}
		line := dy[0]
		s := strings.ReplaceAll(line.Text, "*", "")
		s, _, _ = strings.Cut(s, "[")
		s = strings.TrimSpace(s)
		weekday, dt, ok := strings.Cut(s, " ")
		if !ok && len(days) == 0 {
			err := perror.Error("day-no-day", line.Lineno, "no date for day")
			return err
		}
		weekday = strings.TrimSpace(strings.ToLower(weekday))
		dt = strings.TrimSpace(dt)
		tt := btime.DetectDate(dt)
		if tt == nil {
			err := perror.Error("day-not-valid", line.Lineno, "no valid date `"+dt+"`")
			return err
		}
		if tt.Before(*bdate) {
			err := perror.Error("day-range-from", line.Lineno, "date `"+btime.StringDate(tt, "I")+"` is before "+btime.StringDate(bdate, "I"))
			return err
		}
		if tt.After(*edate) {
			err := perror.Error("day-range-until", line.Lineno, "date `"+dt+"` is after "+btime.StringDate(edate, "I"))
			return err
		}
		ct := btime.StringDate(tt, "D")
		x, _, _ := strings.Cut(ct, " ")
		if x != weekday {
			err := perror.Error("mass-date-day", line.Lineno, "name of day for date `"+dt+"` does not match")
			return err
		}

		day := new(Euday)
		t.Eudays = append(t.Eudays, day)
		day.Date = tt
		day.Start = dy[0].Lineno
		day.Headings = make([]string, 0)
		day.M = make([]*Mass, 0)
		var tm *time.Time = nil

		red := regexp.MustCompile(`^([0-9]{1,2}\.[0-9]{1,2})\s*([a-z]+)([^:]*):(.*)$`)

		for _, line := range dy[1:] {
			s := strings.Replace(line.Text, " u.", " ", 1)
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			pieces := red.FindAllStringSubmatch(s, -1)
			if len(day.M) == 0 && len(pieces) == 0 {
				day.Headings = append(day.Headings, s)
				continue
			}

			if len(pieces) == 0 {
				mass := day.M[len(day.M)-1]
				for _, x := range strings.SplitN(s, ";", -1) {
					y := strings.TrimSpace(x)
					if y == "" {
						continue
					}
					mass.Intentions = append(mass.Intentions, y)
				}
				continue
			}

			hour, min, _ := strings.Cut(pieces[0][1], ".")
			dhour, _ := strconv.Atoi(hour)
			dmin, _ := strconv.Atoi(min)
			if dhour > 25 || dhour < 1 {
				err := perror.Error("mass-hour", line.Lineno, "hour should be 1, ..., 24")
				return err
			}
			if dmin > 59 {
				err := perror.Error("mass-min", line.Lineno, "minutes should be 0, ..., 59")
				return err
			}
			mass := new(Mass)
			day.M = append(day.M, mass)
			tt := time.Date(day.Date.Year(), day.Date.Month(), day.Date.Day(), dhour, dmin, 0, 0, day.Date.Location())
			if tm != nil && tt.Before(*tm) {
				err := perror.Error("mass-seq", line.Lineno, "hour.min is out of sequence")
				return err
			}
			mass.Time = &tt
			tm = mass.Time

			place := pieces[0][2]

			cplace, nplace := findPlace(place)

			if cplace == "" || nplace == "" {
				err := perror.Error("mass-place", line.Lineno, "place `"+place+"` is invalid")
				return err
			}

			mass.Place = cplace

			intention := strings.TrimSpace(pieces[0][4])
			if intention != "" {
				ints := strings.SplitN(intention, ";", -1)
				for _, x := range ints {
					x = strings.TrimSpace(x)
					if x == "EUCHARISTIE" {
						x = "Eucharistie"
					}
					if x != "" {
						x = bstrings.Upper1(x)
						if len(mass.Intentions) != 0 {
							prev := mass.Intentions[len(mass.Intentions)-1]
							if prev == x {
								continue
							}
							if prev == "Eucharistie" {
								mass.Intentions[len(mass.Intentions)-1] = x
								continue
							}
						}
						mass.Intentions = append(mass.Intentions, x)
					}
				}
			}

			players := strings.TrimSpace(pieces[0][3])
			if players != "" {
				roles := strings.SplitN(players, "/", -1)
				if len(roles) > 2 {
					err := perror.Error("mass-role", line.Lineno, "too many roles")
					return err
				}
				for i, role := range roles {
					people := strings.SplitN(role, ";", -1)
					for _, p := range people {
						p := strings.TrimSpace(p)
						if p == "" {
							continue
						}
						cp, np := findPerson(place, p)
						if cp == "" && np == "" {
							err := perror.Error("mass-person", line.Lineno, "person `"+p+"` is invalid")
							return err
						}
						switch i {
						case 0:
							mass.Lectors = append(mass.Lectors, np)
						case 1:
							mass.Dealers = append(mass.Dealers, np)
						}
					}
				}

			}

		}

		t.Body = nil

	}

	doubles := make(map[time.Time]int)
	var last *time.Time = nil
	for _, day := range days {
		start := day.Start
		if doubles[*day.Date] != 0 {
			err := perror.Error("day-double", start, "day occured on line "+strconv.Itoa(doubles[*day.Date]))
			return err
		}
		if last != nil {
			if last.After(*day.Date) {
				err := perror.Error("day-sequence", start, "date out of sequence"+strconv.Itoa(doubles[*day.Date]))
				return err
			}
		}
		last = day.Date
		doubles[*day.Date] = start
	}

	return nil
}

func findPlace(place string) (string, string) {

	placemap := pregistry.Registry["places"].(map[string]any)
	re := regexp.MustCompile(`[^a-zA-Z]`)
	lplace := re.ReplaceAllString(strings.ToLower(place), "")

	found, ok := placemap[lplace]

	if ok {
		x := found.(map[string]any)
		return lplace, x["name"].(string)
	}
	for lp, found := range placemap {
		dx := found.(map[string]any)
		x := dx["name"].(string)
		y := re.ReplaceAllString(strings.ToLower(x), "")
		if y == lplace {
			return lp, x
		}
	}
	return "", ""
}

func findPerson(place string, name string) (string, string) {
	place, _ = findPlace(place)
	if place == "" {
		return "", ""
	}
	placemap := pregistry.Registry["places"].(map[string]any)
	re := regexp.MustCompile(`[^a-zA-Z]`)
	lplace := re.ReplaceAllString(strings.ToLower(place), "")

	f, ok := placemap[lplace]
	if !ok {
		return "", ""
	}
	a := f.(map[string]any)
	b := a["lectors"]
	if b == nil {
		return "", ""
	}
	c := b.(map[string]any)
	if len(c) == 0 {
		return "", ""
	}

	uname := strings.ToUpper(re.ReplaceAllString(name, ""))

	d, ok := c[uname]
	if ok {
		return uname, d.(string)
	}

	for un, f := range c {
		x := f.(string)
		y := strings.ToUpper(re.ReplaceAllString(x, ""))
		if y == uname {
			return un, x
		}
	}
	return "", ""
}
