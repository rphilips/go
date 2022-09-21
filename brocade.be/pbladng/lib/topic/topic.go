package topic

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	pfs "brocade.be/pbladng/lib/fs"
	pimage "brocade.be/pbladng/lib/image"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

var ret = regexp.MustCompile(`^\s*[@#]\s*`)
var reiso = regexp.MustCompile(`^20[0-9][0-9]-[0-9][0-9]$`)
var ws = regexp.MustCompile(`\s`)

type Topic struct {
	Pcodes   []string
	Header   string
	Start    int
	Images   []*pimage.Image
	From     *time.Time
	Until    *time.Time
	LastPB   string
	MaxCount int
	Count    int
	NotePB   string
	NoteMe   string
	Comment  []ptools.Line
	Body     []ptools.Line
	Type     string
	Eudays   []*Euday
}

func (t Topic) String() string {
	J := ptools.J
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n# %s\n", t.Header))
	meta := make([]string, 0)
	if t.Type != "" {
		meta = append(meta, fmt.Sprintf(`"type": %s`, J(t.Type)))
	}
	if len(t.Pcodes) != 0 {
		meta = append(meta, fmt.Sprintf(`"pcodes": %s`, J(strings.Join(t.Pcodes, ";"))))
	}

	if t.From != nil {
		meta = append(meta, fmt.Sprintf(`"from": %s`, J(ptools.StringDate(t.From, "I"))))
	}

	if t.Until != nil {
		meta = append(meta, fmt.Sprintf(`"until": %s`, J(ptools.StringDate(t.Until, "I"))))
	}

	if t.LastPB != "" {
		meta = append(meta, fmt.Sprintf(`"lastpb": %s`, J(t.LastPB)))
	}

	if t.MaxCount != 0 {
		meta = append(meta, fmt.Sprintf(`"maxcount": %s`, J(strconv.Itoa(t.MaxCount))))
	}

	if t.Count != 0 {
		meta = append(meta, fmt.Sprintf(`"count": %s`, J(strconv.Itoa(t.Count))))
	}

	if t.NotePB != "" {
		meta = append(meta, fmt.Sprintf(`"notepb": %s`, J(ptools.Normalize(t.NotePB))))
	}

	if t.NoteMe != "" {
		meta = append(meta, fmt.Sprintf(`"noteme": %s`, J(ptools.Normalize(t.NoteMe))))
	}
	if len(meta) != 0 {
		builder.WriteString("  { ")
		builder.WriteString(strings.Join(meta, ", "))
		builder.WriteString(" }\n")
	}

	if len(t.Comment) != 0 {
		for _, l := range t.Comment {
			builder.WriteString(l.L)
		}
	}
	if len(t.Images) != 0 {
		builder.WriteString("\n")
		for _, img := range t.Images {
			builder.WriteString(img.Name + ".jpg")
			if img.Copyright != "" {
				img.Legend += " © " + img.Copyright
			}
			img.Legend = strings.TrimSpace(img.Legend)
			if img.Legend != "" {
				builder.WriteString(" ")
				builder.WriteString(ptools.Normalize(img.Legend))
				builder.WriteString("\n")
			}
		}
	}
	if len(t.Body) != 0 {
		builder.WriteString("\n")
		for _, line := range t.Body {
			builder.WriteString(ptools.Normalize(line.L))
			builder.WriteString("\n")
		}
	}
	if len(t.Eudays) != 0 {
		for i, euday := range t.Eudays {
			builder.WriteString("\n")
			if i > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(euday.String())
		}
		builder.WriteString("\n")
	}

	return builder.String()
}

func Parse(lines []ptools.Line, mid string, bdate *time.Time, edate *time.Time, checkextern bool) (t *Topic, err error) {
	t = new(Topic)
	for _, line := range lines {
		lineno := line.NR
		s := strings.TrimSpace(line.L)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "//") {
			continue
		}
		nieuw := ret.MatchString(s)
		if !nieuw {
			err = ptools.Error("topic-fluff", lineno, "text outside of topic")
			return
		}
		s = ret.ReplaceAllString(s, "")
		t.Header = ptools.HeaderString(s)
		t.Start = lineno
		break
	}

	if t.Header == "" {
		err = ptools.Error("topic-header", t.Start, "empty topic header")
		return
	}

	if len(lines) == 0 {
		err = ptools.Error("topic-empty", t.Start, "topic without content")
	}
	meta := ptools.Line{}
	images := make([]ptools.Line, 0)
	body := make([]ptools.Line, 0)
	indata := false
	indash := false
	for _, line := range lines {
		s := strings.TrimSpace(line.L)
		if strings.HasPrefix(s, "//") {
			s = strings.TrimPrefix(s, "//")
			if s == "" {
				s = "//"
			} else {
				s = "// " + s
			}
			line.L = s
		}
		if s == "" && !indata {
			continue
		}
		if !indata && strings.HasPrefix(s, "//") {
			t.Comment = append(t.Comment, line)
			continue
		}
		if indata && strings.HasPrefix(s, "//") {
			body = append(body, line)
			continue
		}
		s = strings.ToLower(s)
		if meta.L == "" && strings.HasPrefix(s, "{") {
			meta = line
			continue
		}
		indata = true
		if isdash(s) {
			body = append(body, line)
			indash = !indash
			continue
		}
		if indash {
			body = append(body, line)
			continue
		}
		if strings.Contains(s, ".jpg") {
			images = append(images, line)
			continue
		}
		if line.NR > t.Start {
			body = append(body, line)
		}
	}
	if indash {
		err = ptools.Error("topic-dash", t.Start, "dashline missing")
		return
	}

	if !indata {
		err = ptools.Error("topic-empty", t.Start, "topic without content")
		return
	}

	met := strings.TrimSpace(meta.L)
	if met != "" {
		pcodes, from, until, lastpb, count, maxcount, notepb, noteme, ty, err := JSON(meta)
		if err != nil {
			return nil, err
		}
		t.Pcodes = pcodes
		t.From = from
		t.Until = until
		t.NotePB = notepb
		t.NoteMe = noteme
		t.MaxCount = maxcount
		t.LastPB = lastpb
		t.Count = count
		t.Type = ty
	}
	if t.LastPB < mid {
		t.LastPB = mid
		t.Count += 1
	}

	// dashes

	indash = false
	ok := false
	for _, line := range body {
		s := strings.TrimSpace(line.L)
		if isdash(s) {
			if indash && !ok {
				err = ptools.Error("topic-dash-empty", line.NR, "dash is empty")
				return
			}
			ok = false
			indash = !indash
			continue
		}
		if s != "" && indash {
			ok = true
		}
	}
	if indash {
		err = ptools.Error("topic-dash-open", t.Start, "dash is not closed")
		return
	}

	// images

	if len(images) != 0 {
		dir := pfs.FName("workspace")
		for _, img := range images {
			image, err := pimage.New(img, dir, checkextern)
			if err != nil {
				return nil, err
			}
			t.Images = append(t.Images, &image)
		}
	}

	// body

	t.Body = ptools.WSLines(body)

	for _, line := range t.Body {
		err := ptools.TestLine(line)
		if err != nil {
			return nil, err
		}
	}

	if len(t.Images) == 0 && len(t.Body) == 0 {
		err = ptools.Error("topic-empty2", t.Start, "topic is empty")
		return
	}

	if t.Type == "cal" {
		err = parsecal(t, mid, bdate, edate)
	}

	if t.Type == "mass" {
		err = parseeudays(t, mid, bdate, edate)
	}
	return t, err
}

func JSON(line ptools.Line) (pcodes []string, from *time.Time, until *time.Time, lastpb string, count int, maxcount int, notepb string, noteme string, ty string, err error) {
	met := strings.TrimSpace(line.L)
	lineno := line.NR
	if met == "" {
		return
	}
	if !strings.HasPrefix(met, "{") {
		err = ptools.Error("meta-nojson1", lineno, "should start with `{`")
		return
	}
	if !strings.HasSuffix(met, "}") {
		err = ptools.Error("meta-nojson2", lineno, "should end with `}`")
		return
	}
	mm := make(map[string]string)
	e := json.Unmarshal([]byte(met), &mm)
	if e != nil {
		err = ptools.Error("meta-json-invalid", lineno, "invalid JSON")
		return
	}

	if len(mm) == 0 {
		err = ptools.Error("meta-empty", lineno, "empty meta")
		return
	}
	pcodes = make([]string, 0)
	for key, value := range mm {
		value := strings.TrimSpace(value)
		if value == "" {
			err = ptools.Error("meta-value-empty", lineno, "`"+key+"` is empty")
			return
		}
		switch key {
		case "pcodes":
			validpc := pregistry.Registry["pcodes-valid"].([]string)
			mgood := make(map[string]bool)
			mfound := make(map[string]bool)

			for _, pc := range validpc {
				g := strings.TrimSpace(pc)
				if g != "" {
					mgood[g] = true
				}
			}
			value = strings.ReplaceAll(value, ",", ";")
			given := strings.SplitN(value, ";", -1)
			for _, g := range given {
				g := strings.TrimSpace(g)
				if g == "" {
					err = ptools.Error("meta-pcode-empty", lineno, "empty pcode")
					return
				}
				if !mgood[g] {
					err = ptools.Error("meta-pcode-bad", lineno, "bad pcode")
					return
				}
				if mfound[g] {
					err = ptools.Error("meta-pcode-double", lineno, "pcode `"+g+"` found twice")
					return
				}
				mfound[g] = true
			}
			if len(mfound) == 0 {
				err = ptools.Error("meta-pcode-none", lineno, "no pcodes")
				return
			}
			for pcode := range mfound {
				pcodes = append(pcodes, pcode)
			}

		case "from":
			f, a, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-from-bad", lineno, e)
				return
			}
			if a != "" {
				err = ptools.Error("meta-from-after", lineno, "trailing info after `from`")
				return
			}
			from = f

		case "lastpb":

			if !reiso.MatchString(value) {
				err = ptools.Error("meta-lastpb-bad", lineno, e)
				return
			}
			lastpb = value

		case "count":
			var e error
			count, e = strconv.Atoi(value)
			if e != nil {
				err = ptools.Error("meta-count-bad", lineno, e)
				return
			}
		case "maxcount":
			var e error
			count, e = strconv.Atoi(value)
			if e != nil {
				err = ptools.Error("meta-maxcount-bad", lineno, e)
				return
			}

		case "until":

			u, after, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-until-bad", lineno, e)
				return
			}
			if after != "" {
				err = ptools.Error("meta-until-after", lineno, "trailing info after `until`")
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
				err = ptools.Error("meta-type", lineno, "`"+value+"` is invalid type")
				return
			}
			ty = value

		default:
			err = ptools.Error("meta-key", lineno, "`"+key+"` is unknown")
			return
		}
	}

	return

}

func isdash(s string) bool {
	s = strings.ReplaceAll(s, "_", "-")

	s = ws.ReplaceAllString(s, "")
	if len(s) < 5 {
		return false
	}
	s = strings.TrimLeft(s, "-")
	return s == ""
}
