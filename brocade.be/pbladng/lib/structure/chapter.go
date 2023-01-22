package structure

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	blines "brocade.be/base/lines"
	perror "brocade.be/pbladng/lib/error"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

type Chapter struct {
	Heading  string
	Sort     int
	Topics   []*Topic
	Document *Document
	Until    bool
	Lineno   int
}

func (c *Chapter) LastTopic() (t *Topic) {
	if len(c.Topics) == 0 {
		return
	}
	return c.Topics[len(c.Topics)-1]
}

func (c Chapter) Show() bool {
	for _, t := range c.Topics {
		if t.Show() {
			return true
		}
	}
	return false
}

func (c Chapter) HTML() string {
	if !c.Show() {
		return ""
	}
	builder := strings.Builder{}

	esc := template.HTMLEscapeString
	dash := strings.Repeat("-", 96) + "<br />\n"

	builder.WriteString(fmt.Sprintf("<br /><br /><br />%s%s<br />%s<b>%s</b><br />\n", dash, esc("RUBRIEKTITEL"), dash, esc(c.Heading)))

	for _, topic := range c.Topics {
		builder.WriteString(topic.HTML())
	}
	// builder.WriteString(`<br /><br /><br />`)
	// builder.WriteString(topic)

	builder.WriteString(fmt.Sprintf("<br /><br /><br />%s%s<br />%s", dash, esc("EINDE RUBRIEK"), dash))

	return builder.String()

}

func (c Chapter) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n## %s\n", c.Heading))

	for _, topic := range c.Topics {
		builder.WriteString(topic.String())
	}

	return builder.String()
}

func (c *Chapter) Load(t blines.Text) error {
	tx := blines.Split(t, chexp)
	lineno := tx[1][0].Lineno
	heading := strings.TrimSpace(tx[1][0].Text)
	heading = strings.TrimLeft(heading, ` \t#`)
	heading = ptools.Heading(heading)
	if heading == "" {
		err := perror.Error("chapter-heading-empty", lineno, "empty chapter heading")
		return err
	}
	ok := false
	validti := pregistry.Registry["chapter-heading-regexp"].([]any)
	sortvalue := -1
	wuntil := true
	for i, ti2 := range validti {
		ti := ti2.(map[string]any)["heading"].(string)
		wu := ti2.(map[string]any)["until"].(bool)
		ok = ti == heading
		if !ok {
			rti := ti2.(map[string]any)["regexp"].(string)
			re := regexp.MustCompile(rti)
			ok = re.MatchString(heading)
		}
		if ok {
			sortvalue = i
			heading = ti
			wuntil = wu
			break
		}
	}
	if sortvalue < 0 {
		err := perror.Error("chapter-title-unknown", lineno, fmt.Sprintf("chapter title is invalid: `%s`", heading))
		return err
	}

	c.Heading = heading
	c.Lineno = lineno
	c.Sort = sortvalue
	c.Until = wuntil

	return c.LoadTopics(t)
}

func (c *Chapter) LoadTopics(t blines.Text) error {
	if len(t) < 2 {
		return nil
	}
	t = t[1:]
	ts := blines.Split(t, tpexp)
	first := blines.Compact(ts[0])

	if len(first) != 0 {
		err := perror.Error("chapter-preamble", first[0].Lineno, "chapters should begin with a topic")
		return err
	}

	for i := 1; i < len(ts); i += 2 {
		tc := append(ts[i], ts[i+1]...)
		t := new(Topic)
		t.Chapter = c
		c.Topics = append(c.Topics, t)
		err := t.Load(tc)
		if err != nil {
			return err
		}
	}

	return nil
}
