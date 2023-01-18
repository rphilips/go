package lines

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	bstrings "brocade.be/base/strings"
)

type Line struct {
	Text   string
	Lineno int
}

type Text []Line

func (t Text) String() string {

	data, _ := json.MarshalIndent(t, "", "    ")
	return string(data)

}

func Transform(text Text, fns []func(Text) Text) (t Text) {
	if len(text) == 0 {
		return nil
	}
	if len(fns) == 0 {
		copy(t, text)
		return
	}
	return Transform(fns[0](text), fns[1:])
}

func TrimSpace(text Text) (t Text) {
	for _, l := range text {
		t = append(t, Line{strings.TrimSpace(l.Text), l.Lineno})
	}
	return
}

func LeftTrimSpace(text Text) (t Text) {
	for _, l := range text {
		t = append(t, Line{bstrings.LeftTrimSpace(l.Text), l.Lineno})
	}
	return
}

func RightTrimSpace(text Text) (t Text) {
	for _, l := range text {
		t = append(t, Line{bstrings.RightTrimSpace(l.Text), l.Lineno})
	}
	return
}

func Compact(text Text) (t Text) {
	for _, l := range text {
		line := strings.TrimSpace(l.Text)
		if line != "" {
			t = append(t, Line{line, l.Lineno})
			continue
		}
		if len(t) == 0 {
			continue
		}
		if t[len(t)-1].Text == "" {
			continue
		}
		t = append(t, Line{"", l.Lineno})
	}
	if len(t) == 0 {
		return nil
	}
	if t[len(t)-1].Text == "" {
		return t[:len(t)-1]
	}
	return
}

func ConvertString(body string, start int) (t Text) {
	for i, s := range strings.SplitN(body, "\n", -1) {
		t = append(t, Line{
			Text:   s,
			Lineno: start + i,
		})
	}
	return
}

func ConvertByteSlice(body []byte, start int) (t Text) {
	for i, b := range bytes.SplitN(body, []byte{10}, -1) {
		t = append(t, Line{
			Text:   string(b),
			Lineno: start + i,
		})
	}
	return
}

func Split(t Text, rexp *regexp.Regexp) (ts []Text) {
	if len(t) == 0 {
		return
	}
	ts = append(ts, nil)
	for _, line := range t {
		ok := rexp.MatchString(line.Text)
		if !ok {
			ts[len(ts)-1] = append(ts[len(ts)-1], line)
			continue
		}
		ts = append(ts, Text{line})
		ts = append(ts, nil)
	}
	return
}

func Index(t Text, rexp *regexp.Regexp, start int, end int) (index int) {
	if len(t) == 0 {
		return -1
	}
	if start > -1 && len(t) < start {
		return -1
	}
	if end < 0 || end >= len(t) {
		end = len(t) - 1
	}
	if start < 0 {
		start = 0
	}
	if start > end {
		return -1
	}
	for i := start; i <= end; i++ {
		if rexp.FindStringIndex(t[i].Text) != nil {
			return i
		}
	}
	return -1
}
