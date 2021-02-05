package mfile

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
)

var ws = regexp.MustCompile(`\s`)

// Format formats a M file
func Format(fname string, blob []byte, output *bytes.Buffer) error {

	if blob == nil {
		var err1 error
		blob, err1 = os.ReadFile(fname)
		if err1 != nil {
			e := &qerror.QError{
				Ref:    []string{"mfile.format.read"},
				File:   fname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err1.Error()},
			}
			return e
		}
	}
	blob = qutil.About(blob)
	// check on UTF-8
	body, badutf8, e := qutil.NoUTF8(bytes.NewReader(blob))
	if e != nil {
		err := &qerror.QError{
			Ref:    []string{"mfile.format.utf8"},
			File:   fname,
			Lineno: 1,
			Type:   "Error",
			Msg:    []string{e.Error()},
		}
		return err
	}
	if len(badutf8) != 0 {
		err := &qerror.QError{
			Ref:    []string{"mfile.format.utf8"},
			File:   fname,
			Lineno: badutf8[0][0],
			Type:   "Error",
			Msg:    []string{"Contains non-UTF8"},
		}
		return err
	}
	if len(body) == 0 {
		return nil
	}
	eol := byte('\n')
	stop := false
	blank := true
	about := -1
	buffer := bytes.NewBuffer(body)

	for {
		if stop {
			b := make([]byte, output.Len(), output.Len()+2)
			copy(b, output.Bytes())
			b = bytes.TrimSpace(b)
			output.Reset()
			output.Write(b)
			output.WriteByte(eol)
			return nil
		}
		line, err := buffer.ReadBytes(eol)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF {
			stop = true
			blank = true
		}
		s := strings.TrimSpace(string(line))
		if len(s) != len(line) && (strings.HasPrefix(s, "def ") || strings.HasPrefix(s, "fn ") || strings.HasPrefix(s, "sub ")) {
			line = []byte(s)
		}
		if about == -1 {
			rs := strings.ReplaceAll(s, "/", "")
			if strings.TrimSpace(rs) == "" {
				continue
			}
			about = 0
			if !strings.HasPrefix(s, "/") {
				output.WriteString("// About: \n\n")
				about = 1
				blank = true
			}
		}
		if about == 0 {
			rs := strings.TrimSpace(strings.ReplaceAll(s, "/", ""))
			if blank && rs == "" {
				continue
			}
			if rs == "" {
				output.WriteString("\n")
				blank = true
				continue
			}
			if strings.HasPrefix(s, "/") {
				output.WriteString(s + "\n")
				blank = false
				continue
			}
			about = 1
			blank = true
		}

		if blank && s == "" {
			continue
		}

		if s == "" {
			blank = true
			output.WriteString("\n")
			continue
		}

		// real M

		// dotted
		dots := ""
		next := -1
		if strings.HasPrefix(s, ".") {
			for k, ch := range s {
				if ch == '.' {
					dots += "."
					continue
				}
				if unicode.IsSpace(ch) {
					continue
				}
				next = k
				break
			}

			if next == -1 {
				s = ""
			} else {
				s = s[next:]
			}
		}
		if dots != "" {
			output.WriteString(" " + dots + " " + s + "\n")
			blank = false
			continue
		}
		if strings.HasPrefix(s, "/") {
			output.WriteString(s + "\n")
			blank = false
			continue
		}
		if strings.HasPrefix(s, ";") {
			output.WriteString(" " + s + "\n")
			blank = false
			continue
		}
		sub := strings.HasPrefix(s, "def") || strings.HasPrefix(s, "sub") || strings.HasPrefix(s, "fn") || strings.HasPrefix(s, "%")
		prefix := ""
		sr := strings.TrimRightFunc(string(line), unicode.IsSpace)
		if len(sr) != len(s) && !sub {
			prefix = " "
		}

		if sub && !blank {
			output.WriteString("\n")
			blank = true
		}

		if sub {
			switch {
			case strings.HasPrefix(s, "def"):
				s = strings.TrimSpace(s[3:])
				prefix = "def "
			case strings.HasPrefix(s, "sub"):
				s = strings.TrimSpace(s[3:])
				prefix = "sub "
			case strings.HasPrefix(s, "fn"):
				s = strings.TrimSpace(s[2:])
				prefix = "fn "
			}
		}

		s = ws.ReplaceAllString(s, " ")
		if prefix == " " {
			output.WriteString(" " + s + "\n")
			blank = false
			continue
		}
		output.WriteString(prefix + s + "\n")
		blank = false
		continue
	}
}
