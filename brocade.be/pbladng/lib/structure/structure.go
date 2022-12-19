package structure

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	btime "brocade.be/base/time"
	perror "brocade.be/pbladng/lib/error"
)

type Document struct {
	Year     int
	Week     int
	Bdate    *time.Time
	Edate    *time.Time
	Mailed   *time.Time
	Chapters []*Chapter
}

type Manifest struct {
	Letter string `json:"letter"`
	FName  string `json:"fname"`
	Digest string `json:"digest"`
}

type Chapter struct {
	Heading  string
	Sort     int
	Topics   []*Topic
	Document *Document
	Lineno   int
}

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
	Comment  []string
	Body     []string
	Eudays   []*Euday
	Images   []*Image
	Chapter  *Chapter
	Lineno   int
}

type Euday struct {
	Date     *time.Time
	Headings []string
	Masses   []*Mass
}

type Mass struct {
	Time       *time.Time
	Place      string
	Lectors    []string
	Dealers    []string
	Intentions []string
}

type Image struct {
	Name      string
	Legend    string
	Copyright string
	Fname     string
	Lineno    int
}

type ImageID struct {
	Path   string `json:"path"`
	Mtime  string `json:"mtime"`
	Digest string `json:"digest"`
	Letter string `json:"letter"`
}

func (d *Document) LastChapter() (c *Chapter) {
	if len(d.Chapters) == 0 {
		return
	}
	return d.Chapters[len(d.Chapters)-1]
}

func (d *Document) LastTopic() (t *Topic) {
	c := d.LastChapter()
	if c == nil {
		return
	}
	return c.LastTopic()
}

func (d *Document) Test(hint string) (err error) {
	if d.Year == 0 {
		err = perror.Error("document-missing-meta", 1, "meta information is missing [hint="+hint+"]")
		return
	}
	return
}

func (c *Chapter) LastTopic() (t *Topic) {
	if len(c.Topics) == 0 {
		return
	}
	return c.Topics[len(c.Topics)-1]
}

func (c *Chapter) Test(hint string) (err error) {
	if c.Heading == "" {
		err = perror.Error("chapter-empty-title", c.Lineno, "chapter has no title [hint="+hint+"]")
		return
	}
	return
}

func (t *Topic) Test(hint string) (err error) {

	if t.Heading == "" {
		err = perror.Error("chapter-empty-title", t.Lineno, "chapter has no title [hint="+hint+"]")
		return
	}
	return
}

func (d Document) ID() string {
	return fmt.Sprintf("%d-%02d", d.Year, d.Week)
}

func (d Document) String() string {
	m := make(map[string]string)
	m["id"] = d.ID()
	m["bdate"] = btime.StringDate(d.Bdate, "I")
	m["edate"] = btime.StringDate(d.Edate, "I")
	if d.Mailed != nil {
		m["mailed"] = btime.StringDate(d.Mailed, "I")
	}
	j, _ := json.MarshalIndent(m, "", "    ")

	builder := strings.Builder{}

	builder.Write(j)
	builder.WriteString("\n")
	for _, chapter := range d.Chapters {
		builder.WriteString(chapter.String())
	}
	return builder.String()
}

func (c Chapter) String() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("\n\n\n# %s\n", c.Heading))
	// if c.Text != "" {
	// 	builder.WriteString(c.Text)
	// } else {
	// 	for _, topic := range c.Topics {
	// 		builder.WriteString(topic.String())
	// 	}
	// }
	return builder.String()
}
