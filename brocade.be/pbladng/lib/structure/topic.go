package structure

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	blines "brocade.be/base/lines"
	bstrings "brocade.be/base/strings"
	btime "brocade.be/base/time"
	perror "brocade.be/pbladng/lib/error"
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

func (t Topic) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n# %s\n", t.Heading))
	meta := make([]string, 0)
	if t.Type != "" {
		meta = append(meta, fmt.Sprintf(`"type": %s`, bstrings.JSON(t.Type)))
	}
	if t.From != nil {
		meta = append(meta, fmt.Sprintf(`"from": %s`, bstrings.JSON(ptools.StringDate(t.From, "I"))))
	}

	if t.Until != nil {
		meta = append(meta, fmt.Sprintf(`"until": %s`, bstrings.JSON(ptools.StringDate(t.Until, "I"))))
	}

	if t.LastPB != "" {
		meta = append(meta, fmt.Sprintf(`"lastpb": %s`, bstrings.JSON(t.LastPB)))
	}

	if t.MaxCount != 0 {
		meta = append(meta, fmt.Sprintf(`"maxcount": %s`, bstrings.JSON(strconv.Itoa(t.MaxCount))))
	}

	if t.Count != 0 {
		meta = append(meta, fmt.Sprintf(`"count": %s`, bstrings.JSON(strconv.Itoa(t.Count))))
	}

	if t.NotePB != "" {
		meta = append(meta, fmt.Sprintf(`"notepb": %s`, bstrings.JSON(ptools.Normalize(t.NotePB, true))))
	}

	if t.NoteMe != "" {
		meta = append(meta, fmt.Sprintf(`"noteme": %s`, bstrings.JSON(ptools.Normalize(t.NoteMe, true))))
	}
	if len(meta) != 0 {
		builder.WriteString("  { ")
		builder.WriteString(strings.Join(meta, ", "))
		builder.WriteString(" }\n")
	}

	if len(t.Images) != 0 {
		builder.WriteString("\n")
		for _, img := range t.Images {
			fmt.Printf("%v\n", img)
			builder.WriteString(img.Name + ".jpg")
			if img.Copyright != "" {
				img.Legend += " © " + img.Copyright
			}
			img.Legend = strings.TrimSpace(img.Legend)
			if img.Legend != "" {
				builder.WriteString(" ")
				builder.WriteString(ptools.Normalize(img.Legend, true))
				builder.WriteString("\n")
			}
		}
	}
	if len(t.Body) != 0 {
		builder.WriteString("\n")
		for _, line := range t.Body {
			builder.WriteString(ptools.Normalize(line.Text, true))
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
		if value == "" {
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
			t.NotePB = value

		case "noteme":
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
