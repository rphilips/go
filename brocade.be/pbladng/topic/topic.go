package topic

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	ptools "brocade.be/pbladng/tools"
)

type Image struct {
	Name      string
	Header    string
	Copyright string
	Fname     string
	Lineno    int
}

type Topic struct {
	Header string
	Images []*Image
	From   string
	Until  string
	Note   string
	Body   string
	Lineno int
}

func (Topic) New(body []byte, lineno int) (*Topic, error) {
	err := ptools.IsUTF8(body, lineno)
	if err != nil {
		return nil, err
	}
	body, extra := ptools.LeftStrip(body)
	if len(body) == 0 {
		err := fmt.Errorf("error line %d: empty topic", lineno)
		return nil, err
	}
	lineno += extra
	sbody := string(body)

	header, from, until, note, rest, err := Header(sbody, lineno)
	if err != nil {
		return nil, err
	}
	topic := Topic{
		Header: header,
		From:   from,
		Until:  until,
		Note:   note,
		Lineno: lineno,
	}

	brest, extra := ptools.LeftStrip([]byte(rest))
	rest = string(brest)
	lineno += extra

	images, extra, rest, err := Images(rest, lineno)
	if err != nil {
		return nil, err
	}
	topic.Images = images
	lineno += extra

	topic.Body, err = Parse(rest, lineno)

	return &topic, err
}

func Header(body string, lineno int) (header string, from string, until string, note string, rest string, err error) {
	if !strings.HasPrefix(body, "#") {
		err = fmt.Errorf("error line %d: header should start with `#`", lineno)
		return
	}
	lines := strings.SplitN(body, "\n", -1)
	header = strings.TrimSpace(lines[0][1:])
	if header == "" {
		err = fmt.Errorf("error line %d: header should not be empty", lineno)
		return
	}
	rest = strings.Join(lines[1:], "\n")
	if !strings.Contains(header, "[") {
		return
	}
	k := strings.LastIndex(header, "[")
	work := strings.TrimSpace(header[k+1:])
	header = strings.TrimSpace(header[:k])
	if header == "" {
		err = fmt.Errorf("error line %d: header should not be empty", lineno)
		return
	}
	if !strings.HasSuffix(work, "]") {
		err = fmt.Errorf("error line %d: header should end on `]`", lineno)
		return
	}
	work = strings.TrimSuffix(work, "]")
	work = strings.TrimSpace(work)
	parts := strings.SplitN(work, ";", -1)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		before, after := ptools.LeftWord(part)
		before = strings.ToLower(before)

		switch before {
		case "from":
			if from != "" {
				err = fmt.Errorf("error line %d: `from` twice defined", lineno)
				return
			}
			from = after
		case "until":
			if until != "" {
				err = fmt.Errorf("error line %d: `until` twice defined", lineno)
				return
			}
			until = after
		case "note":
			if until != "" {
				err = fmt.Errorf("error line %d: `note` twice defined", lineno)
				return
			}
			note = after
		default:
			err = fmt.Errorf("error line %d: should start with `note; from; until`", lineno)
			return
		}
	}
	if from != "" {
		x, e := strconv.Atoi(from)
		if e == nil {
			if x > 53 {
				err = fmt.Errorf("error line %d: *from* should be smaller than 53", lineno)
				return
			}
			if x < 1 {
				err = fmt.Errorf("error line %d: *from* should not be smaller than 1", lineno)
				return
			}
		}
		if e != nil {
			_, err = ptools.ParseIsoDate(from)
			err = fmt.Errorf("error line %d: %s", lineno, err)
		}
	}
	if until != "" {
		x, e := strconv.Atoi(until)
		if e == nil {
			if x > 53 {
				err = fmt.Errorf("error line %d: *until* should be smaller than 53", lineno)
				return
			}
			if x < 1 {
				err = fmt.Errorf("error line %d: *until* should not be smaller than 1", lineno)
				return
			}
		}
		if e != nil {
			_, err = ptools.ParseIsoDate(until)
			err = fmt.Errorf("error line %d: %s", lineno, err)
		}
	}

	return
}

func Images(body string, lineno int) (images []*Image, extra int, rest string, err error) {
	lines := strings.SplitN(body, "\n", -1)
	re := regexp.MustCompile(`\.[Jj][Pp][Ee]?[Gg]`)
	reimg := regexp.MustCompile(`^[a-zA-Z0-9_]+\.jpg`)
	recopy := regexp.MustCompile(`[cC][Oo][Pp][Yy][Rr][Ii][Gg][Hh][Tt]`)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if re.MatchString(line) && !strings.Contains(line, ".jpg") {
			err = fmt.Errorf("error line %d: image extensions should be `.jpg`", lineno+i)
			return
		}
		if !strings.Contains(line, ".jpg") {
			rest = strings.Join(lines[i:], "\n")
			extra = i
			break
		}
		before, after := ptools.FirstAlfa(line)
		before = strings.TrimSpace(before)
		if before != "" {
			err = fmt.Errorf("error line %d: image name should start at the beginning of line`", lineno+i)
			return
		}
		after = strings.TrimSpace(after)
		if !reimg.MatchString(after) {
			err = fmt.Errorf("error line %d: image name should be of the form [a-zA-Z0-9_]+.jpg`", lineno+i)
			return
		}
		parts := strings.SplitN(after, ".jpg", 2)
		img := new(Image)
		img.Name = parts[0]
		text := strings.TrimSpace(parts[1])
		text = recopy.ReplaceAllString(text, "©")
		img.Header = text
		if strings.ContainsRune(text, '©') {
			parts := strings.SplitN(text, "©", 2)
			img.Header = strings.TrimSpace(parts[0])
			img.Copyright = strings.TrimSpace(parts[1])
		}
		images = append(images, img)
	}
	if re.MatchString(rest) {
		lines := strings.SplitN(rest, "\n", -1)
		for i, line := range lines {
			if re.MatchString(line) {
				err = fmt.Errorf("error line %d: images should stand at the beginning of the topic", lineno+extra+i)
				return
			}
		}
	}
	return
}

func Parse(body string, lineno int) (bulk string, err error) {
	// Euro transformation
	body = strings.ReplaceAll(body, "€", " EUR ")
	body = ptools.Euro(body)
	// number handling
	keep := ""
	for {
		before, number, after := ptools.NumberSplit(body)
		if number == -1 {
			break
		}
		if before != "" {
			r, _ := utf8.DecodeLastRuneInString(before)
			if unicode.IsLetter(r) {
				keep += before + strconv.Itoa(number)
				body = after
				continue
			}
		}
	}

	return
}
