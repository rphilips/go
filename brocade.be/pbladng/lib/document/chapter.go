package document

import (
	"fmt"
	"regexp"

	perror "brocade.be/pbladng/lib/error"
	pregistry "brocade.be/pbladng/lib/registry"
)

func ChapterHeading(s string, lineno int) (r string, sort int, err error) {
	validtis := pregistry.Registry["chapter-heading-regexp"].([]any)
	sort = -1
	for i, ti2 := range validtis {
		ti := ti2.(map[string]any)["regexp"].(string)
		re := regexp.MustCompile(ti)
		if re.MatchString(s) {
			sort = i
			r = ti2.(map[string]any)["heading"].(string)
			break
		}
	}
	if sort == -1 {
		err = perror.Error("chapter-heading-unknown", lineno, fmt.Sprintf("chapter without known heading `%s`", s))
		return
	}
	return
}
