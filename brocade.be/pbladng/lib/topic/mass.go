package topic

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

type Euday struct {
	Date    *time.Time
	Headers []string
	Start   int
	M       []*Mass
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
	weekday, _, _ := strings.Cut(ptools.StringDate(d.Date, "D"), " ")
	weekday = strings.ToUpper(weekday[0:1]) + weekday[1:]

	headers := strings.Join(d.Headers, "\n")
	headers = strings.TrimSpace(headers)

	if headers != "" {
		headers = headers + "\n"
	}

	mas := ""
	for _, m := range d.M {
		mas += strings.TrimSpace(m.String()) + "\n"
	}
	return strings.TrimSpace(fmt.Sprintf("%s %02d/%02d\n%s%s", weekday, day, month, headers, mas))
}

func (d Euday) HTML() string {
	weekday := ptools.StringDate(d.Date, "D")
	weekday = strings.ToUpper(weekday[0:1]) + weekday[1:]
	esc := template.HTMLEscapeString
	h := ptools.Html
	headers := strings.Join(d.Headers, "\n")
	headers = strings.TrimSpace(headers)
	if headers != "" {
		headers = h(esc(headers))
		headers = strings.ReplaceAll(headers, "\n", "<br />")
		headers += "<br />"
	}
	mas := ""
	for _, m := range d.M {
		mas += m.HTML() + "<br />"
	}
	return strings.TrimSpace(fmt.Sprintf("<b>%s</b><br />%s%s", weekday, headers, mas))
}

func parseeudays(topic *Topic, mid string, bdate *time.Time, edate *time.Time) (err error) {

	dys := make([][]ptools.Line, 0)
	days := make([]*Euday, 0)
	rem1 := regexp.MustCompile(`^[A-Z][a-z]+dag [0-9]+ [a-z]+ [0-9]{4}$`)
	rem2 := regexp.MustCompile(`^[A-Z][a-z]+dag [0-9]{1,2}/[0-9]{1,2}$`)
	for _, line := range topic.Body {
		s := strings.ReplaceAll(line.L, "*", "")
		s, _, _ = strings.Cut(s, "[")
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		nieuw := rem1.MatchString(s) || rem2.MatchString(s)
		if !nieuw && len(dys) == 0 {
			err = ptools.Error("day-fluff", line.NR, "text outside of day")
			return
		}
		if nieuw {
			dys = append(dys, make([]ptools.Line, 0))
		}
		dys[len(dys)-1] = append(dys[len(dys)-1], line)
	}
	if len(dys) == 0 {
		err = ptools.Error("day-missing", topic.Start, "mass missing")
		return
	}
	// per dag
	for _, dy := range dys {
		if len(dy) < 2 {
			err = ptools.Error("day-missing", dy[0].NR, "info missing")
			return
		}
		line := dy[0]
		s := strings.ReplaceAll(line.L, "*", "")
		s, _, _ = strings.Cut(s, "[")
		s = strings.TrimSpace(s)
		weekday, dt, ok := strings.Cut(s, " ")
		if !ok && len(days) == 0 {
			err = ptools.Error("day-no-day", line.NR, "no date for day")
			return
		}
		weekday = strings.TrimSpace(strings.ToLower(weekday))
		dt = strings.TrimSpace(dt)
		t, after, e := ptools.NewDate(dt)
		if e != nil {
			err = ptools.Error("day-not-valid", line.NR, "no valid date `"+dt+"`")
			return
		}
		if strings.TrimSpace(after) != "" {
			err = ptools.Error("day-after", line.NR, "superfluous text after date")
			return
		}
		if t.Before(*bdate) {
			err = ptools.Error("day-range-from", line.NR, "date `"+ptools.StringDate(t, "I")+"` is before "+ptools.StringDate(bdate, "I"))
			return
		}
		if t.After(*edate) {
			err = ptools.Error("day-range-until", line.NR, "date `"+dt+"` is after "+ptools.StringDate(edate, "I"))
			return
		}
		ct := ptools.StringDate(t, "D")
		x, _, _ := strings.Cut(ct, " ")
		if x != weekday {
			err = ptools.Error("mass-date-day", line.NR, "name of day for date `"+dt+"` does not match")
			return
		}

		day := new(Euday)
		topic.Eudays = append(topic.Eudays, day)
		day.Date = t
		day.Start = dy[0].NR
		day.Headers = make([]string, 0)
		day.M = make([]*Mass, 0)
		var tm *time.Time = nil

		red := regexp.MustCompile(`^([0-9]{1,2}\.[0-9]{1,2})\s*([a-z]+)([^:]*):(.*)$`)

		for _, line := range dy[1:] {
			s := strings.Replace(line.L, " u.", " ", 1)
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			pieces := red.FindAllStringSubmatch(s, -1)
			if len(day.M) == 0 && len(pieces) == 0 {
				day.Headers = append(day.Headers, s)
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
				err = ptools.Error("mass-hour", line.NR, "hour should be 1, ..., 24")
				return
			}
			if dmin > 59 {
				err = ptools.Error("mass-min", line.NR, "minutes should be 0, ..., 59")
				return
			}
			mass := new(Mass)
			day.M = append(day.M, mass)
			tt := time.Date(day.Date.Year(), day.Date.Month(), day.Date.Day(), dhour, dmin, 0, 0, day.Date.Location())
			if tm != nil && tt.Before(*tm) {
				err = ptools.Error("mass-seq", line.NR, "hour.min is out of sequence")
				return
			}
			mass.Time = &tt
			tm = mass.Time

			place := pieces[0][2]

			cplace, nplace := findPlace(place)

			if cplace == "" || nplace == "" {
				err = ptools.Error("mass-place", line.NR, "place `"+place+"` is invalid")
				return
			}

			mass.Place = cplace

			intention := strings.TrimSpace(pieces[0][4])
			if intention != "" {
				ints := strings.SplitN(intention, ";", -1)
				for _, x := range ints {
					x = strings.TrimSpace(x)
					if x != "" {
						mass.Intentions = append(mass.Intentions, x)
					}
				}
			}

			players := strings.TrimSpace(pieces[0][3])
			if players != "" {
				roles := strings.SplitN(players, "/", -1)
				if len(roles) > 2 {
					err = ptools.Error("mass-role", line.NR, "too many roles")
					return
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
							err = ptools.Error("mass-person", line.NR, "person `"+p+"` is invalid")
							return
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

		topic.Body = nil

	}

	doubles := make(map[time.Time]int)
	var last *time.Time = nil
	for _, day := range days {
		start := day.Start
		if doubles[*day.Date] != 0 {
			err = ptools.Error("day-double", start, "day occured on line "+strconv.Itoa(doubles[*day.Date]))
			return
		}
		if last != nil {
			if last.After(*day.Date) {
				err = ptools.Error("day-sequence", start, "date out of sequence"+strconv.Itoa(doubles[*day.Date]))
				return
			}
		}
		last = day.Date
		doubles[*day.Date] = start
	}
	return
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
