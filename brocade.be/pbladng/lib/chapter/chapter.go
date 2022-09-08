package chapter

import (
	"fmt"
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
}

func (c Chapter) String() string {
	builder := strings.Builder{}
	for i, topic := range c.Topics {
		if i == 0 {
			builder.WriteString(fmt.Sprintf("\n\n\n## %s\n", ptools.HeaderString(c.Header)))

		}
		builder.WriteString(topic.String())
	}
	return builder.String()
}

func Parse(lines []ptools.Line, mid string, bdate *time.Time, edate *time.Time) (c *Chapter, err error) {
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
		ti := (ti2.([]any))[0]
		re := regexp.MustCompile(ti.(string))
		ok = re.MatchString(c.Header)
		if ok {
			sortvalue = i
			break
		}
	}
	if !ok {
		err = ptools.Error("chapter-title-unknown", c.Start, fmt.Sprintf("chapter without unknown title `%s`", c.Header))
		return
	}
	c.Sort = sortvalue

	tops := make([][]ptools.Line, 0)

	for _, line := range lines {
		if line.NR <= c.Start {
			continue
		}
		s := strings.TrimSpace(line.L)
		if s == "" && (len(tops) == 0 || len(tops[len(tops)-1]) == 0) {
			continue
		}

		nieuw := ret.MatchString(line.L)
		if s != "" && len(tops) == 0 && !nieuw {
			err = ptools.Error("topics-fluff", line.NR, "text outside of topic")
			return
		}
		if nieuw {
			tops = append(tops, make([]ptools.Line, 0))
		}
		tops[len(tops)-1] = append(tops[len(tops)-1], line)
	}
	if len(tops) == 0 {
		return
	}

	for _, top := range tops {
		topic, e := ptopic.Parse(top, mid, bdate, edate)
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
