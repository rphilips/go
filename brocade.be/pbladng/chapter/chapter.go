package chapter

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	pregistry "brocade.be/pbladng/registry"
	ptools "brocade.be/pbladng/tools"
	ptopic "brocade.be/pbladng/topic"
)

type Chapter struct {
	Sort   int
	Header string
	Start  int
	Topics []*ptopic.Topic
}

func Parse(lines []ptools.Line) (chapters []*Chapter, err error) {
	chaps := make([][]ptools.Line, 0)
	chapters = make([]*Chapter, 0)
	for _, line := range lines {
		s := strings.TrimSpace(line.L)
		if !strings.HasPrefix(s, "#") {
			if len(chaps) != 0 {
				chaps[len(chaps)-1] = append(chaps[len(chaps)-1], line)
			}
			continue
		}
		s = strings.TrimSpace(s[1:])
		if strings.HasPrefix(s, "#") {
			chaps = append(chaps, make([]ptools.Line, 0))
		}
		chaps[len(chaps)-1] = append(chaps[len(chaps)-1], line)
	}
	if len(chaps) == 0 {
		return
	}
	for _, chap := range chaps {
		chapter, err := One(chap)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, chapter)
	}
	if len(chapters) != 0 {
		sort.Slice(chapters, func(i, j int) bool { return chapters[i].Sort < chapters[j].Sort })
	}
	return
}

func One(chap []ptools.Line) (chapter *Chapter, err error) {
	line := chap[0]
	s := strings.TrimSpace(line.L)
	s = s[1:]
	s = strings.TrimSpace(s)
	s = s[1:]
	s = strings.TrimSpace(s)
	lineno := line.NR
	if s == "" {
		err = ptools.Error("chapter-empty", lineno, "chapter without title")
	}
	s = strings.ToUpper(s)
	ok := false
	validti := make([]string, 0)
	json.Unmarshal([]byte(pregistry.Registry["chapter-title-regex"]), &validti)
	sortvalue := -1
	for i, ti := range validti {
		re := regexp.MustCompile(ti)
		ok = re.MatchString(s)
		if ok {
			sortvalue = i
			break
		}
	}
	if !ok {
		err = ptools.Error("chapter-title-unknown", lineno, fmt.Sprintf("chapter without unknown title `%s`", s))
	}
	chapter = new(Chapter)
	chapter.Sort = sortvalue
	chapter.Header = s
	for _, line := range chap[1:] {
		s := strings.TrimSpace(line.L)
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "#") {
			break
		}
		return nil, ptools.Error("chapter-fluff ", lineno, "non-empty chapter should start with `#`")
	}
	topics, err := ptopic.Parse(chap[1:])

	if err != nil {
		chapter.Topics = topics
	}
	return chapter, err
}
