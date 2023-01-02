package lines

import (
	"bytes"
	"regexp"
	"strings"

	bstrings "brocade.be/base/strings"
)

type Line struct {
	Text   string
	Lineno int
}

type Text []Line

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

func ConvertString(body string) (t Text) {
	for i, s := range strings.SplitN(body, "\n", -1) {
		t = append(t, Line{
			Text:   s,
			Lineno: 1 + i,
		})
	}
	return
}

func ConvertByteSlice(body []byte) (t Text) {
	for i, b := range bytes.SplitN(body, []byte{10}, -1) {
		t = append(t, Line{
			Text:   string(b),
			Lineno: 1 + i,
		})
	}
	return
}

func Split(t Text, rexp regexp.Regexp) (ts []Text) {
	for _, line := range t {
		ok := rexp.MatchString(line.Text)
		if !ok {
			if len(ts) == 0 {
				ts = append(ts, Text{line})
			} else {
				ts[len(ts)-1] = append(ts[len(ts)-1], line)
			}
			continue
		}
		ts = append(ts, Text{line})
		ts = append(ts, nil)
	}
	return
}
