package topic

import (
	"regexp"
	"time"
)

var ren *regexp.Regexp = regexp.MustCompile(`\d+`)

func parsecal(topic *Topic, mid string, bdate *time.Time, edate *time.Time) (err error) {

	return nil
}
