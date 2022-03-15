package util

import (
	"bytes"
	"io"
	"strings"
)

func CommentPartition(blob []byte) (comment []byte, body []byte, about []byte) {
	if len(blob) == 0 {
		return nil, blob, nil
	}
	delim1 := []byte("//")
	delim2 := []byte(`"""`)
	delim3 := []byte(`'''`)
	delims := [][]byte{ delim1, delim2, delim3 }
	var delim []byte = nil
	empty = []byte("")
	utf8 = []byte("-*- coding: utf-8 -*-")
	abot = []byte("About:")

	eol := []byte("\n")
	lines := bytes.SplitN(blob, eol, -1)
	// search start of comment
	start := -1
	for i, line := range lines {
		b = bytes.ReplaceAll(line, utf8, empty)
		t := bytes.TrimSpace(b)
		if len(t) == 0 {
			continue
		}
		for _, d := range delims {
			if !bytes.HasPrefix(t, delim) {
				continue
			}
			delim = d
		}
		if delim != nil {
			start = i
		}
		break
	}
	if delim == nil {
		return nil, blob, nil
	}
	// Handle comment section

	if !bytes.Equal(delim, delim1) {
		t := bytes.TrimSpace(lines[count:])
		lines[start] = t[3:]
	}
	end := -1

	for i, line := range lines[start:] {
		end = i
		if !bytes.Equal(delim, delim1) {
			if bytes.Contains(line, delim) {
				break
			}
			continue
		}
		b = bytes.ReplaceAll(line, utf8, empty)
		b = bytes.ReplaceAll(b, abot, empty)
		t := bytes.TrimSpace(b)
		if len(t) == 0 {
			continue
		}
		if !bytes.HasPrefix(t, delim) {
			end--
			break
		}
	}

	if end < count {
		return nil, blob, nil
	}

	cmt := make([][]byte, 0)
	for _, line := range lines[start:end+1] {
		b = bytes.ReplaceAll(line, utf8, empty)
		b = bytes.ReplaceAll(b, abot, empty)
		t := bytes.TrimSpace(b)
		if bytes.Equal(delim, delim1) {
			t = bytes.TrimLeft(t, "/")
			line = t
		}
		if len(cmt) == 0 && len(t) == 0 {
			continue
		}
		if len(cmt) == 0 {
			about = t
			cmt = append(cmt, append([]byte("// About: ", t...)))
			continue
		}



	}






		if bytes.HasPrefix(t, delim2) {
			count = i
			continue
		}
		if bytes.HasPrefix(t, delim3) {
			count = i
			continue
		}

			&& !bytes.HasPrefix(line, delim2)

	}

}

// About changes the comment form to '//'
func About(blob []byte) (result []byte) {
	buffer := bytes.NewBuffer(blob)
	comment := make([]string, 0)
	body := make([]byte, 0)
	eol := byte('\n')
	delim := ""
	stop := false
	for {
		if stop {
			break
		}
		if delim == "??" {
			body = append(body, buffer.Bytes()...)
			break
		}

		line, err := buffer.ReadBytes(eol)
		if err != nil && err != io.EOF {
			return blob[:]
		}
		stop = err == io.EOF
		s := string(line)
		s = strings.ReplaceAll(s, "-*- coding: utf-8 -*-", "")
		s = strings.ReplaceAll(s, "About:", " ")
		t := strings.TrimSpace(s)
		if delim == "//" && !strings.HasPrefix(t, "//") {
			body = append(body, line...)
			delim = "??"
			continue
		}
		if delim == "" {
			if t == "" {
				continue
			}
			if strings.HasPrefix(t, "//") {
				delim = "//"
			}
			if delim == "" && strings.HasPrefix(t, "'''") {
				delim = "'''"
				s = strings.TrimPrefix(t, delim)
				t = strings.TrimSpace(s)
			}
			if delim == "" && strings.HasPrefix(t, `"""`) {
				delim = `"""`
				s = strings.TrimPrefix(t, delim)
				t = strings.TrimSpace(s)
			}
		}
		if delim == "" {
			return blob[:]
		}
		if delim != "//" && strings.Contains(t, delim) {
			s = strings.SplitN(s, delim, 2)[0]
			delim = "??"
		}
		comment = append(comment, s)
	}
	cmt := make([]string, 0)
	for _, s := range comment {
		s = RStrip(s)
		t := strings.TrimSpace(s)
		if strings.HasPrefix(t, "//") {
			s = strings.TrimLeft(t, "/")
			s = RStrip(s)
			if len(cmt) == 0 && s == "" {
				continue
			}
			cmt = append(cmt, strings.TrimSpace(s))
			continue
		}
		if len(cmt) == 0 && t == "" {
			continue
		}
		if len(cmt) == 0 {
			s = t
		}
		if len(cmt) != 0 && strings.TrimSpace(cmt[len(cmt)-1]) == t {
			continue
		}
		cmt = append(cmt, s)
	}
	if len(cmt) == 0 {
		return body
	}
	cmt[0] = "About: " + cmt[0]
	for i, s := range cmt {
		if s == "" {
			cmt[i] = "//"
		} else {
			if strings.HasPrefix(s, " ") {
				cmt[i] = "//" + s
			} else {
				cmt[i] = "// " + s
			}
		}
	}
	scmt := strings.TrimSpace(strings.Join(cmt, "\n")) + "\n\n"
	body = append(bytes.TrimSpace(body), 10)
	result = append([]byte(scmt), body...)
	return result
}

// AboutLine retrieves the first About line
func AboutLine(blob []byte) string {
	buffer := bytes.NewBuffer(blob)
	eol := byte('\n')
	slash := []byte("//")
	for {
		line, err := buffer.ReadBytes(eol)
		if err != nil && err != io.EOF {
			return ""
		}
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		if !bytes.HasPrefix(line, slash) {
			return ""
		}
		sline := string(line)
		k := strings.Index(sline, "About:")
		if k < 0 {
			continue
		}
		sline = sline[k+6:]
		return strings.TrimSpace(sline)
	}
}

// Comment maakt een commentaar lijn uit een slice of strings
func Comment(cmt interface{}) string {
	if cmt == nil {
		return ""
	}
	c := make([]string, 0)
	lines := strings.SplitN(cmt.(string), "\n", -1)
	if len(lines) == 0 {
		return ""
	}

	found := -1
	for _, l := range lines {
		l = strings.TrimRight(l, "\t\r ")
		l = strings.TrimLeft(l, " \t")
		l2 := strings.TrimLeft(l, "/")
		if l2 == "" && len(c) == 0 {
			continue
		}
		if l2 == "" {
			l = "//"
		}
		if len(l) > 2 && l[2:3] != " " {
			l = "// " + l[2:]
		}
		if len(c) == 0 {
			c = append(c, l)
			if len(l) > 3 {
				found = len(c)
			}
			continue
		}
		if c[len(c)-1] == l {
			continue
		}
		c = append(c, l)
		if len(l) > 3 {
			found = len(c)
		}
	}
	if found == -1 {
		return ""
	}
	return strings.TrimSpace(strings.Join(c[:found], "\n"))
}
