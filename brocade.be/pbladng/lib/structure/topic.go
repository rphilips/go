package structure

import (
	"encoding/json"
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

var rjpg = regexp.MustCompile(`\.[Jj][Pp][Ee]?[Gg]`)
var reiso = regexp.MustCompile(`^20[0-9][0-9]-[0-9][0-9]$`)

type Topic struct {
	Type     string
	Heading  string
	From     *time.Time
	Until    *time.Time
	LastPB   string
	MaxCount int
	Count    int
	NotePB   string
	NoteMe   string
	Body     blines.Text
	Eudays   []*Euday
	Images   []*Image
	Chapter  *Chapter
	Lineno   int
}

func (t Topic) Show() bool {
	c := t.Chapter
	doc := c.Document
	bdate := doc.Bdate
	if t.From != nil && !t.From.Before(*bdate) {
		return false
	}
	if t.Type == "mass" {
		return true
	}
	if len(t.Images) != 0 {
		return true
	}
	for _, line := range t.Body {
		s := strings.TrimSpace(line.Text)
		s = strings.TrimLeft(s, " \t=")
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "//") {
			continue
		}
		return true
	}
	return false
}
func (t Topic) HTML() string {
	if !t.Show() {
		return ""
	}
	c := t.Chapter
	doc := c.Document
	week := doc.Week
	letters := doc.Letters
	builder := strings.Builder{}

	esc := template.HTMLEscapeString
	dash := strings.Repeat("-", 96) + "<br />\n"
	h := ptools.Html

	if t.NotePB != "" {
		builder.WriteString(fmt.Sprintf("<br /><br /><br />%sBEGIN TOPIC: %s<br />%s", dash, h(esc(t.NotePB)), dash))
	}
	builder.WriteString(fmt.Sprintf("<br /><br /><br /><b>%s</b><br /><br />", h(esc(ptools.HeadingString(t.Heading)))))

	// builder.WriteString(`<br /><br /><br />`)
	// builder.WriteString(topic)
	// images
	alfabet := "abcdefghijklmnopqrstuvwxyz"
	if len(t.Images) > 0 {
		for _, img := range t.Images {
			builder.WriteString(dash)
			legend := img.Legend
			cr := img.Copyright
			legend += " © " + cr
			legend = strings.TrimSpace(legend)
			legend = strings.TrimRight(legend, "©")
			if legend != "" {
				legend = "\u00A0" + legend
			}
			letter := 1 + len(letters)
			if len(alfabet) < letter {
				continue
			}
			imgletter := alfabet[letter-1 : letter]
			letters += imgletter
			builder.WriteString(fmt.Sprintf("F%s%s%02d.jpg%s<br />\n", esc(pregistry.Registry["pcode"].(string)), imgletter, week, h(esc(legend))))
			builder.WriteString(dash)
		}
		doc.Letters = letters
	}

	first := true

	for _, line := range t.Body {
		text := line.Text
		if strings.HasPrefix(text, "//") {
			continue
		}

		text = h(esc(text))
		if first && text == "" {
			continue
		}
		if first {
			builder.WriteString("<br />")
			first = false
		}
		builder.WriteString(text)
		builder.WriteString("<br />")
	}

	if len(t.Eudays) != 0 {
		for i, euday := range t.Eudays {
			x := euday.HTML()
			if x == "" {
				continue
			}
			if first {
				builder.WriteString("<br />")
				first = false
			}
			builder.WriteString(x)
			if len(t.Eudays) != i+1 {
				builder.WriteString("<br />")
			}
		}
	}

	if t.NotePB != "" {
		builder.WriteString(fmt.Sprintf("%sEINDE TOPIC<br />%s", dash, dash))
	}
	return builder.String()

}

func (t Topic) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n# %s\n", t.Heading))
	meta := make([]string, 0)
	if t.Type != "" || true {
		meta = append(meta, fmt.Sprintf(`"type": %s`, bstrings.JSON(t.Type)))
	}

	if t.Until != nil || true {
		meta = append(meta, fmt.Sprintf(`"until": %s`, bstrings.JSON(btime.StringDate(t.Until, "I"))))
	}

	if t.NotePB != "" || true {
		s, _ := ptools.Normalize(t.NotePB, true)
		meta = append(meta, fmt.Sprintf(`"notepb": %s`, bstrings.JSON(s)))
	}

	if t.From != nil || true {
		meta = append(meta, fmt.Sprintf(`"from": %s`, bstrings.JSON(btime.StringDate(t.From, "I"))))
	}

	if t.LastPB != "" || true {
		meta = append(meta, fmt.Sprintf(`"lastpb": %s`, bstrings.JSON(t.LastPB)))
	}

	if t.MaxCount != 0 || true {
		meta = append(meta, fmt.Sprintf(`"maxcount": %s`, bstrings.JSON(strconv.Itoa(t.MaxCount))))
	}

	if t.Count != 0 || true {
		meta = append(meta, fmt.Sprintf(`"count": %s`, bstrings.JSON(strconv.Itoa(t.Count))))
	}

	if t.NoteMe != "" || true {
		s, _ := ptools.Normalize(t.NoteMe, true)
		meta = append(meta, fmt.Sprintf(`"noteme": %s`, bstrings.JSON(s)))
	}
	if len(meta) != 0 {
		builder.WriteString("  { ")
		builder.WriteString(strings.Join(meta, ", "))
		builder.WriteString(" }\n")
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
				s, _ := ptools.Normalize(img.Legend, true)
				builder.WriteString(s)
				builder.WriteString("\n")
			}
		}
	}
	maxday := ""
	if len(t.Body) != 0 {
		builder.WriteString("\n")
		for _, line := range t.Body {
			s, md := ptools.Normalize(line.Text, true)
			if md > maxday {
				maxday = md
			}
			builder.WriteString(s)
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

func (t *Topic) Load(ts blines.Text) error {
	tx := blines.Split(ts, tpexp)

	lineno := tx[1][0].Lineno
	heading := strings.TrimSpace(tx[1][0].Text)
	heading = strings.TrimLeft(heading, ` \t#`)
	heading = ptools.Heading(heading)
	if heading == "" {
		err := perror.Error("topic-heading-empty", lineno, "empty topic heading")
		return err
	}
	set := `\|*_`
	s := heading
	if strings.ContainsAny(s, set) {
		for i, r := range set {
			if !strings.ContainsRune(s, r) {
				continue
			}
			rs := string(byte(i))
			s = strings.ReplaceAll(s, `\`+string(r), rs)
		}
		if strings.Contains(s, `*`) {
			err := perror.Error("topic-heading-*", lineno, "heading should not contain unescaped `*`: "+s)
			return err
		}
		if strings.Contains(s, `_`) {
			err := perror.Error("topic-heading-_", lineno, "heading should not contain unescaped `_`")
			return err
		}
	}

	ts = blines.Compact(tx[2])
	lineno = tx[2][0].Lineno
	if len(ts) == 0 {
		return perror.Error("topic-empty", lineno, "topic should not be empty")
	}
	ts, err := t.LoadMeta(ts)
	if err != nil {
		return err
	}
	ts, err = t.LoadImages(ts)
	if err != nil {
		return err
	}
	t.Heading = heading
	t.Body = blines.Compact(ts)
	t.Lineno = lineno
	err = t.LoadCal()
	if err != nil {
		return err
	}
	err = t.LoadMass()

	if t.Until == nil && t.Chapter.Until {
		c := t.Chapter
		doc := c.Document
		t.Until = doc.Bdate
		//return perror.Error("topic-until", lineno, "`until` is missing")
	}
	maxday := ""

	for i := 0; i < len(t.Body); i++ {
		md := ""
		s := t.Body[i].Text
		if !strings.HasPrefix(s, "=") {
			err := ptools.CheckIBAN(s)
			if err != nil {
				return perror.Error("topic-iban", t.Body[i].Lineno, err.Error())
			}
		}
		t.Body[i].Text, md = ptools.Normalize(t.Body[i].Text, true)
		if md > maxday {
			maxday = md
		}
	}

	if maxday != "" {
		if t.Until == nil {
			t.Until = btime.DetectDate(maxday)
		} else {
			uday := btime.StringDate(t.Until, "I")
			if !strings.HasPrefix(uday, "3") {
				t.Until = btime.DetectDate(maxday)
			}
		}
	}

	return err
}

func (t *Topic) LoadMeta(tx blines.Text) (txo blines.Text, err error) {
	first := tx[0]
	if !strings.HasPrefix(first.Text, "{") {
		return tx, nil
	}
	slicemeta := make([]string, 0)
	ok := false
	lineno1 := 0
	lineno2 := 0

	for i, line := range tx {
		slicemeta = append(slicemeta, line.Text)
		if !strings.HasSuffix(first.Text, "}") {
			continue
		}
		ok = true
		lineno2 = i
		break
	}

	if !ok {
		return tx, perror.Error("meta-close", first.Lineno, "meta does not have a close `}`")
	}
	txo = tx[lineno1+1:]

	smeta := strings.Join(slicemeta, "\n")

	findLineno := func(key string) int {
		rexp := regexp.MustCompile(`"\s*` + regexp.QuoteMeta(key) + `\s*"\s*:`)
		index := blines.Index(tx, rexp, lineno1, lineno2)
		if index < 0 {
			return -1
		}
		return tx[index].Lineno
	}
	mm := make(map[string]string)
	e := json.Unmarshal([]byte(smeta), &mm)
	if e != nil {
		err = perror.Error("meta-json-invalid", first.Lineno, "invalid JSON")
		return txo, err
	}

	if len(mm) == 0 {
		err = perror.Error("meta-empty", first.Lineno, "empty meta")
		return txo, err
	}
	for key, value := range mm {
		value := strings.TrimSpace(value)
		if false && value == "" {
			lineno := findLineno(key)
			err = perror.Error("meta-value-empty", lineno, "`"+key+"` is empty")
			return txo, err
		}
		switch key {
		case "from":
			f := btime.DetectDate(value)
			if f == nil {
				lineno := findLineno(key)
				err = perror.Error("meta-from-bad", lineno, e)
				return txo, err
			}
			t.From = f

		case "lastpb":

			if !reiso.MatchString(value) {
				lineno := findLineno(key)
				err = perror.Error("meta-lastpb-bad", lineno, e)
				return txo, err
			}
			t.LastPB = value

		case "count":
			var e error
			count, e := strconv.Atoi(value)
			if e != nil {
				lineno := findLineno(key)
				err = perror.Error("meta-count-bad", lineno, e)
				return txo, err
			}
			t.Count = count

		case "maxcount":
			var e error
			count, e := strconv.Atoi(value)
			if e != nil {
				lineno := findLineno(key)
				err = perror.Error("meta-maxcount-bad", lineno, e)
				return txo, err
			}
			t.MaxCount = count

		case "until":
			u := btime.DetectDate(value)
			if u == nil {
				lineno := findLineno(key)
				err = perror.Error("meta-until-bad", lineno, e)
				return txo, err
			}
			t.Until = u

		case "notepb":
			if strings.Contains(value, "\n") {
				lineno := findLineno(key)
				err = perror.Error("meta-notepb-multiple", lineno, "note to pblad should be a oneliner")
				return txo, err
			}
			value, _ := ptools.Normalize(value, true)
			t.NotePB = value

		case "noteme":
			value, _ := ptools.Normalize(value, true)
			t.NoteMe = value

		case "type":
			value = strings.ToLower(value)
			if value != "cal" && value != "mass" {
				lineno := findLineno(key)
				err = perror.Error("meta-type", lineno, "`"+value+"` is invalid type")
				return txo, err
			}
			t.Type = value
		default:
			lineno := findLineno(key)
			err = perror.Error("meta-key", lineno, "`"+key+"` is invalid")
			return txo, err
		}
	}

	return txo, nil
}

func (t *Topic) LoadImages(tx blines.Text) (txo blines.Text, err error) {
	doc := t.Chapter.Document
	pdir := doc.ArchiveDirPrevious()
	dirs := []string{doc.Dir, pdir}
	copyright := ""
	for _, line := range tx {
		s := line.Text
		if strings.HasPrefix(s, "#") {
			err = perror.Error("topic-loose", line.Lineno, "line starting with #")
			return tx, err
		}
		if strings.HasPrefix(s, "//") {
			txo = append(txo, line)
			continue
		}
		for _, r := range "|_*" {
			s = ptools.DoubleChar(s, r)
		}
		line.Text = s
		bal := ptools.CheckBalanced(s)
		if bal != "" {
			err = perror.Error("topic-balance", line.Lineno, "unbalanced characters `"+bal+"`")
			return tx, err
		}

		if rjpg.FindStringIndex(s) != nil {
			img, err := NewImage(line.Text, copyright, line.Lineno, dirs)
			if err != nil {
				return tx, err
			}
			copyright = img.Copyright
			t.Images = append(t.Images, img)
			continue
		}
		txo = append(txo, line)
		continue
	}
	txo = blines.Compact(txo)
	return txo, nil
}

func (t *Topic) LoadCal() error {
	if t.Type != "cal" {
		return nil
	}
	return nil
}
