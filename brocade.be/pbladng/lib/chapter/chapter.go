package chapter

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
	"time"

	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
	ptopic "brocade.be/pbladng/lib/topic"
)

var rec = regexp.MustCompile(`^\s*[@#]\s*[@#]`)
var ret = regexp.MustCompile(`^\s*[@#]`)

type Chapter struct {
	Sort   int
	Header string
	Start  int
	Topics []*ptopic.Topic
	Text   string
}

func (c Chapter) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n## %s\n", ptools.HeaderString(c.Header)))
	if c.Text != "" {
		builder.WriteString(c.Text)
	} else {
		for _, topic := range c.Topics {
			builder.WriteString(topic.String())
		}
	}
	return builder.String()
}

func (c Chapter) HTML(bdate, edate *time.Time, id string, imgletters map[string]string) string {
	topics := make([]string, 0)
	for _, topic := range c.Topics {
		ht := topic.HTML(bdate, edate, id, imgletters)
		if ht == "" {
			continue
		}
		topics = append(topics, ht)
	}
	if len(topics) == 0 {
		return ""
	}

	builder := strings.Builder{}
	esc := template.HTMLEscapeString
	dash := strings.Repeat("-", 96) + "<br />\n"

	builder.WriteString(fmt.Sprintf("<br /><br /><br />%s%s<br />%s<b>%s</b><br />\n", dash, esc("RUBRIEKTITEL"), dash, esc(ptools.HeaderString(c.Header))))

	for _, topic := range topics {
		builder.WriteString(`<br /><br /><br />`)
		builder.WriteString(topic)
	}

	builder.WriteString(fmt.Sprintf("<br /><br /><br />%s%s<br />%s", dash, esc("EINDE RUBRIEK"), dash))

	return builder.String()
}

func Parse(lines []ptools.Line, mid string, bdate *time.Time, edate *time.Time, checkextern bool) (c *Chapter, err error) {
	c = new(Chapter)
	for _, line := range lines {
		lineno := line.NR
		s := strings.TrimSpace(line.L)
		if s == "" {
			continue
		}
		nieuw := rec.MatchString(s)
		if !nieuw {
			err = ptools.Error("chapter-fluff2", lineno, "text outside of chapter")
			return
		}
		s = rec.ReplaceAllString(s, "")
		c.Header = ptools.HeaderString(s)
		c.Start = lineno
		break
	}

	if c.Header == "" {
		err = ptools.Error("chapter-header", c.Start, "empty chapter header")
		return
	}

	ok := false
	validti := pregistry.Registry["chapter-title-regexp"].([]any)
	sortvalue := -1
	for i, ti2 := range validti {
		ti := ti2.(map[string]any)["regexp"].(string)
		re := regexp.MustCompile(ti)
		ok = re.MatchString(c.Header)
		if ok {
			sortvalue = i
			break
		}
	}
	if !ok {
		err = ptools.Error("chapter-title-unknown", c.Start, fmt.Sprintf("chapter without known title `%s`", c.Header))
		return
	}
	c.Sort = sortvalue

	tops := make([][]ptools.Line, 0)

	for _, line := range lines {
		if line.NR <= c.Start {
			continue
		}
		s := strings.TrimSpace(line.L)
		switch {
		case s == "" && len(tops) != 0:
			tops[len(tops)-1] = append(tops[len(tops)-1], line)
			continue
		case s == "":
			continue
		default:
			nieuw := ret.MatchString(line.L)
			switch {
			case nieuw:
				tops = append(tops, make([]ptools.Line, 0))
				tops[len(tops)-1] = append(tops[len(tops)-1], line)
				continue
			case len(tops) == 0:
				err = ptools.Error("topics-fluff", line.NR, "text outside of topic: "+s)
				return
			default:
				tops[len(tops)-1] = append(tops[len(tops)-1], line)
				continue
			}
		}
	}
	if len(tops) == 0 {
		return
	}

	for _, top := range tops {
		topic, e := ptopic.Parse(top, mid, bdate, edate, checkextern)
		if e != nil {
			return c, e
		}
		c.Topics = append(c.Topics, topic)
	}

	doubles := make(map[string]int)
	for _, t := range c.Topics {
		ti := ptools.HeaderString(t.Header)
		k := strings.Index(ti, "[")
		if k != -1 {
			ti = strings.TrimSpace(ti[:k])
		}
		if doubles[ti] != 0 {
			err = ptools.Error("topic-double", t.Start, "title occured on line "+strconv.Itoa(doubles[ti]))
			return
		}
		doubles[ti] = t.Start
	}
	return
}
