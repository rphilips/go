package document

import (
	"fmt"
	"strings"

	perror "brocade.be/pbladng/lib/error"
	pstructure "brocade.be/pbladng/lib/structure"
	ptools "brocade.be/pbladng/lib/tools"
)

func TopicHeading(s string, lineno int) (r string, err error) {
	r = strings.TrimSpace(s)

	if r == "" {
		err = perror.Error("topic-heading-empty", lineno, "topic has empty heading")
		return
	}
	return
}

func AddTopic(doc *pstructure.Document, content *string) (err error) {
	c := doc.LastChapter()
	if c == nil {
		return fmt.Errorf("no chapter defined yet")
	}
	t := c.LastTopic()
	if t == nil {
		return fmt.Errorf("no topic defined yet")
	}

	blob := strings.TrimSpace(*content)
	if blob == "" {
		err = ptools.Error("topic-empty", t.Lineno, "topic without content")
		return
	}

	lines := strings.SplitN(strings.TrimSpace(*content), "\n", -1)
	body := make([]string, 0, len(lines))
	indata := false
	indash := false
	for _, line := range lines {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "//") {
			s = strings.TrimPrefix(s, "//")
			if s == "" {
				s = "//"
			} else {
				s = "// " + s
			}
			line = s
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
		indata = true
		if ptools.Isdash(s) {
			body = append(body, line)
			indash = !indash
			continue
		}
		if indash {
			body = append(body, line)
			continue
		}
	}
	if indash {
		err = ptools.Error("topic-dash", t.Lineno, "dashline missing")
		return
	}
	t.Body = ptools.WSLines(body)
	return nil
}
