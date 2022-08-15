package topic

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	pimage "brocade.be/pbladng/image"
	pregistry "brocade.be/pbladng/registry"
	ptools "brocade.be/pbladng/tools"
)

type Topic struct {
	Pcodes []string
	Header string
	Images []*pimage.Image
	From   *time.Time
	Until  *time.Time
	Note   string
	Body   []ptools.Line
}

func Parse(lines []ptools.Line) (topics []*Topic, err error) {
	tops := make([][]ptools.Line, 0)
	topics = make([]*Topic, 0)
	for _, line := range lines {
		s := strings.TrimSpace(line.L)
		if strings.HasPrefix(s, "#") {
			tops = append(tops, make([]ptools.Line, 0))
		}
		tops[len(tops)-1] = append(tops[len(tops)-1], line)
	}
	if len(tops) == 0 {
		return
	}
	for _, top := range tops {
		topic, err := One(top)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return
}

func One(top []ptools.Line) (topic *Topic, err error) {
	line := top[0]
	s := strings.TrimSpace(line.L)
	s = s[1:]
	s = strings.TrimSpace(s)
	lineno := line.NR
	if s == "" {
		err = ptools.Error("topic-title-empty", lineno, "topic without title")
		return
	}
	meta := ptools.Line{}
	topic = new(Topic)
	topic.Header = strings.ToUpper(s)
	images := make([]ptools.Line, 0)
	body := make([]ptools.Line, 0)
	indata := false
	indash := false
	for _, line := range top[1:] {
		s := strings.TrimSpace(line.L)
		if s == "" && !indata {
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
		if strings.Index(s, ".jpg") == -1 {
			images = append(images, line)
			continue
		}
		body = append(body, line)
	}
	if indash {
		err = ptools.Error("topic-dash", lineno, "dashline missing")
		return
	}

	if !indata {
		err = ptools.Error("topic-empty", lineno, "topic without content")
		return
	}

	met := strings.TrimSpace(meta.L)
	if met != "" {
		pcodes, from, until, note, err := JSON(meta)
		if err != nil {
			return nil, err
		}
		topic.Pcodes = pcodes
		topic.From = from
		topic.Until = until
		topic.Note = note
	}

	if len(images) != 0 {
		dir := pregistry.Registry["workspace-path"]
		for _, img := range images {
			image, err := pimage.New(img, dir)
			if err != nil {
				return nil, err
			}
			topic.Images = append(topic.Images, &image)
		}
	}
	topic.Body = body
	return
}

func JSON(line ptools.Line) (pcodes []string, from *time.Time, until *time.Time, note string, err error) {
	met := strings.TrimSpace(line.L)
	lineno := line.NR
	if met == "" {
		return
	}
	if !strings.HasPrefix(met, "{") {
		err = ptools.Error("meta-nojson", lineno, "should start with `{`")
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
		switch key {
		case "pcodes":
			validpc := make([]string, 0)
			json.Unmarshal([]byte(pregistry.Registry["pcodes-valid"]), &validpc)
			mgood := make(map[string]bool)
			mfound := make(map[string]bool)

			for _, pc := range validpc {
				g := strings.TrimSpace(pc)
				if g != "" {
					mgood[g] = true
				}
			}
			given := strings.SplitN(value, ",", -1)
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
			if value == "" {
				err = ptools.Error("meta-from-empty", lineno, "`from` is empty")
				return
			}

			f, a, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-from-bad", lineno, e.Error())
				return
			}
			if a != "" {
				err = ptools.Error("meta-from-after", lineno, "trailing info after `from`")
				return
			}
			from = f

		case "until":
			if value == "" {
				err = ptools.Error("meta-until-empty", lineno, "`until` is empty")
				return
			}

			u, after, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-until-bad", lineno, e.Error())
				return
			}
			if after != "" {
				err = ptools.Error("meta-until-after", lineno, "trailing info after `until`")
				return
			}
			until = u

		case "note":
			if value == "" {
				err = ptools.Error("meta-note-empty", lineno, "`note` is empty")
				return
			}
			note = value

		default:
			err = ptools.Error("meta-key", lineno, "`"+key+"` is unknown")
			return
		}
	}

	return

}

func isvalidmap(mm map[string]string, lineno int) (good []string, from *time.Time, until *time.Time, note string, err error) {
	if len(mm) == 0 {
		err = ptools.Error("meta-empty", lineno, "empty meta")
		return
	}
	good = make([]string, 0)
	for key, value := range mm {
		value := strings.TrimSpace(value)
		switch key {
		case "pcodes":

			validpc := make([]string, 0)
			json.Unmarshal([]byte(pregistry.Registry["pcodes-valid"]), &validpc)
			pcodes := strings.SplitN(value, ",", -1)
			for _, pcode := range pcodes {
				ok := false
				pcode := strings.TrimSpace(pcode)
				if pcode == "" {
					err = ptools.Error("meta-pcode-empty", lineno, "lege pcode")
					return
				}
				for _, pc := range validpc {
					ok = pcode == pc
					if ok {
						good = append(good, pc)
						break
					}
				}
				if !ok {
					err = ptools.Error("meta-pcode-bad", lineno, "bad pcode")
					return
				}
			}
			if len(good) == 0 {
				err = ptools.Error("meta-pcode-none", lineno, "no pcodes")
				return
			}
		case "from":
			if value == "" {
				err = ptools.Error("meta-from-empty", lineno, "`from` is empty")
				return
			}

			f, a, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-from-bad", lineno, e.Error())
				return
			}
			if a != "" {
				err = ptools.Error("meta-from-after", lineno, "trailing info after `from`")
				return
			}
			from = f
		case "until":
			if value == "" {
				err = ptools.Error("meta-until-empty", lineno, "`until` is empty")
				return
			}

			u, after, e := ptools.NewDate(value)
			if e != nil {
				err = ptools.Error("meta-until-bad", lineno, e.Error())
				return
			}
			if after != "" {
				err = ptools.Error("meta-until-after", lineno, "trailing info after `until`")
				return
			}
			until = u
		case "note":
			if value == "" {
				err = ptools.Error("meta-note-empty", lineno, "`note` is empty")
				return
			}
			note = value
		default:
			err = ptools.Error("meta-key", lineno, "`"+key+"` is unknown")
			return
		}
	}

	return

}

func isdash(s string) bool {
	s = strings.ReplaceAll(s, "_", "-")
	re := regexp.MustCompile(`\s`)
	s = re.ReplaceAllString(s, "")
	if len(s) < 5 {
		return false
	}
	s = strings.TrimLeft(s, "-")
	return s == ""
}
