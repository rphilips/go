package document

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
)

// Retrieve year and week without parsing
func DocRef(dir string) (year int, week int, mailed string, err error) {
	if dir == "" {
		d, ok := pregistry.Registry["distribute-dir"]

		if !ok || d.(string) == "" {
			dir = pfs.FName("workspace")
		} else {
			dir = d.(string)
		}

	}
	weekpb := ""
	for _, base := range []string{"week.md", "parochieblad.ed", "week.pb"} {
		weekpb = filepath.Join(dir, base)
		if bfs.Exists(weekpb) {
			break
		}
	}
	if weekpb == "" {
		err = fmt.Errorf("cannot find week.[md,pb]")
		return
	}

	data, err := os.ReadFile(weekpb)
	if err != nil {
		err = fmt.Errorf("cannot read `%s`: %s", weekpb, err.Error())
		return
	}

	value := ""
	if strings.HasPrefix(string(data), "WEEK") {
		s := strings.TrimPrefix(string(data), "WEEK")
		s = strings.TrimSpace(s)
		value, _, _ = strings.Cut(s, " ")
		jmail := filepath.Join(dir, "mailed.ok")
		data, e := os.ReadFile(jmail)
		if e == nil {
			m := make(map[string]string)
			json.Unmarshal(data, &m)
			mailed = m["nazareth"]
		}
	}
	if value == "" {

		_, a, ok := strings.Cut(string(data), "{")

		if !ok {
			err = fmt.Errorf("`%s` does not contain `{`", weekpb)
			return
		}

		m, _, ok := strings.Cut(a, "}")
		if !ok {
			err = fmt.Errorf("`%s` does not contain `}` after first `{`", weekpb)
			return
		}

		m = "{" + m + "}"

		meta := make(map[string]string)
		err = json.Unmarshal([]byte(m), &meta)
		if err != nil {
			err = fmt.Errorf("`%s` does not contain valid JSON between first `{` and `}`", weekpb)
			return
		}

		value = meta["id"]
		mailed = meta["mailed"]
	}
	y, w, ok := strings.Cut(value, "-")
	if !ok {
		if y == "" || w == "" {
			err = fmt.Errorf("`id` is missing in `%s", weekpb)
			return
		}
	}
	year, e := strconv.Atoi(y)
	if e != nil {
		err = fmt.Errorf("id `%s` should start with a valid year in `%s", value, weekpb)
		return
	}
	week, e = strconv.Atoi(w)
	if e != nil {
		err = fmt.Errorf("id `%s` should end with a valid week in `%s", value, weekpb)
		return
	}
	return
}

// func Parse(reader io.Reader, dir string) (doc *pstructure.Document, codes map[string]bool, alts map[string]*pstructure.Image, err error) {
// 	blob, linenos, err := IsUTF8(reader, true)
// 	if err != nil {
// 		return
// 	}
// 	r := mtext.NewReader(blob)
// 	md := mgold.New()
// 	parser := md.Parser()
// 	docnode := parser.Parse(r)
// 	doc = new(pstructure.Document)
// 	start := 0
// 	content := ""
// 	mode := ""
// 	until := 0
// 	codes = make(map[string]bool)
// 	alts = make(map[string]*pstructure.Image)
// 	err = parse(doc, docnode, blob, nil, &start, &content, &mode, &until, linenos, codes, alts, dir)
// 	if err != nil {
// 		codes = nil
// 		alts = nil
// 		return
// 	}
// 	return
// }

// func parse(doc *pstructure.Document, n mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int, codes map[string]bool, alts map[string]*pstructure.Image, dir string) (err error) {
// 	kind := n.Kind()
// 	name := kind.String()
// 	ty := n.Type()
// 	sty := "inline"
// 	if ty == mast.TypeBlock {
// 		sty = "block"
// 	}
// 	switch name {
// 	case "Document":
// 	case "TextBlock", "ThematicBreak", "CodeBlock", "Blockquote", "List", "ListItem", "HTMLBlock":
// 		lineno := 0
// 		if n.Lines().Len() != 0 {
// 			line := n.Lines().At(0)
// 			*start = line.Start
// 		}
// 		if lineno == 0 && *start != 0 {
// 			lineno = Lineno(*start, linenos)
// 		}
// 		err = perror.Error("parse-"+sty, lineno, fmt.Sprintf("`%s` is not allowed", name))
// 	case "Link", "AutoLink", "RawHTML", "String":
// 		lineno := 0
// 		if *start != 0 {
// 			lineno = Lineno(*start, linenos)
// 		}
// 		err = perror.Error("parse-"+sty, lineno, fmt.Sprintf("`%s` is not allowed", name))
// 	case "Heading":
// 		err = AstHeading(doc, n, blob, pn, start, content, mode, until, linenos)
// 	case "FencedCodeBlock":
// 		err = AstFencedCodeBlock(doc, n, blob, pn, start, content, mode, until, linenos)
// 	case "Paragraph":
// 		err = AstParagraph(doc, n, blob, pn, start, content, mode, until, linenos)
// 	case "Emphasis":
// 		err = AstEmphasis(doc, n, blob, pn, start, content, mode, until, linenos)
// 	case "Image":
// 		err = AstImage(doc, n, blob, pn, start, content, linenos, alts, dir)
// 	case "Text":
// 		err = AstText(doc, n, blob, pn, start, content, mode, until, linenos, alts, dir)
// 	case "CodeSpan":
// 		code, err := AstCodeSpan(doc, n, blob, pn, start, content, mode, until, linenos)
// 		if err == nil {
// 			codes[code] = true
// 		}
// 	default:
// 		err = perror.Error("parse-"+sty, *start, fmt.Sprintf("`%s` is not allowed", name))
// 		return
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	if name == "FencedCodeBlock" && dir == "" {
// 		return nil
// 	}

// 	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
// 		err := parse(doc, c, blob, pn, start, content, mode, until, linenos, codes, alts, dir)
// 		pn = c
// 		if err != nil {
// 			return err
// 		}
// 		if dir == "" {
// 			return nil
// 		}

// 	}
// 	return err
// }

// func AstHeading(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int) (err error) {
// 	n := node.(*mast.Heading)
// 	level := n.Level
// 	if n.Lines().Len() == 0 {
// 		lineno := 0
// 		if *start != 0 {
// 			rest := string(blob[*start:])
// 			k := strings.Index(rest, "#")
// 			rest = rest[:k]
// 			lineno = Lineno(*start, linenos)
// 			lineno += strings.Count(rest, "\n")
// 		}
// 		err = perror.Error("md-invalid-heading", lineno, "invalid heading")
// 		return
// 	}

// 	line := node.Lines().At(0)
// 	lno := Lineno(line.Start, linenos)

// 	heading := ptools.Heading(strings.TrimSpace(string(Content(n, blob))))

// 	if heading == "" {
// 		err = perror.Error("doc-heading-empty", lno, "heading is empty")
// 		return
// 	}

// 	if err == nil {
// 		switch level {
// 		case 1:
// 			r, sort, e := ChapterHeading(heading, lno)
// 			if e != nil {
// 				err = e
// 				return
// 			}
// 			c := new(pstructure.Chapter)
// 			c.Lineno = lno
// 			c.Document = doc
// 			c.Heading = r
// 			c.Sort = sort
// 			doc.Chapters = append(doc.Chapters, c)
// 		case 2:
// 			c := doc.LastChapter()
// 			if c == nil {
// 				err = perror.Error("topic-missing-chapter", lno, "topic before chapter")
// 				break
// 			}
// 			if err != nil {
// 				break
// 			}
// 			r, e := TopicHeading(heading, lno)
// 			if e != nil {
// 				err = e
// 				return
// 			}
// 			t := new(pstructure.Topic)
// 			t.Heading = r
// 			t.Chapter = c
// 			t.Lineno = lno
// 			c.Topics = append(c.Topics, t)

// 		default:
// 			err = perror.Error("topic-sublevel", lno, "topics do not have sublevels")
// 		}

// 	}
// 	if err == nil {
// 		*start = line.Stop
// 		*content = ""
// 		*mode = "heading"
// 		*until = line.Stop
// 	}
// 	return
// }

// func AstFencedCodeBlock(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int) (err error) {
// 	n := node.(*mast.FencedCodeBlock)
// 	lineno := 0
// 	if n.Lines().Len() == 0 {
// 		if *start != 0 {
// 			lineno = Lineno(*start, linenos)
// 		}
// 		err = perror.Error("md-invalid-fence", lineno, "invalid fence")
// 		return
// 	}
// 	line := node.Lines().At(0)
// 	lno := Lineno(line.Start, linenos)
// 	*start = line.Stop
// 	if strings.TrimSpace(*content) != "" {
// 		err = perror.Error("code-prefixed", lno, "JSON should not be prefixed with text: ["+*content+"]")
// 		return
// 	}

// 	//pnm := node.Kind().String()
// 	language := string(n.Language(blob))
// 	if language != "json" {
// 		err = perror.Error("code-json", lno, "only JSON is accepted")
// 		return
// 	}
// 	js := Content(node, blob)

// 	meta := make(map[string]string)

// 	e := json.Unmarshal(js, &meta)
// 	if e != nil {
// 		return perror.Error("json-unmarshal", lno, e)
// 	}

// 	switch doc.Year {
// 	case 0:
// 		c := doc.LastChapter()
// 		if c != nil {
// 			err = perror.Error("doc-meta-first", lno, "document meta should come first")
// 			return
// 		}
// 		year, week, bdate, edate, mailed, e := DocJSON(meta, lno)
// 		if e != nil {
// 			err = e
// 			return
// 		}
// 		doc.Year = year
// 		doc.Week = week
// 		doc.Bdate = bdate
// 		doc.Edate = edate
// 		doc.Mailed = mailed
// 	default:
// 		t := doc.LastTopic()
// 		if t == nil {
// 			err = perror.Error("topic-meta-place1", lno, "topic meta should come after heading")
// 			return
// 		}
// 		if len(t.Eudays) != 0 || len(t.Body) != 0 || len(t.Images) != 0 {
// 			err = perror.Error("topic-meta-place2", lno, "topic meta should come before rest of topic")
// 			return
// 		}
// 		if t.From != nil || t.Until != nil || t.Type != "" || t.NotePB != "" || t.NoteMe != "" || t.MaxCount != 0 || t.LastPB != "" || t.Count != 0 {
// 			err = perror.Error("topic-meta-extra", lno, "topic meta should be defined only once")
// 			return
// 		}
// 		from, until, lastpb, count, maxcount, notepb, noteme, ty, e := TopicJSON(meta, *start)
// 		if e != nil {
// 			err = e
// 			return
// 		}
// 		t.From = from
// 		t.Until = until
// 		t.NotePB = notepb
// 		t.NoteMe = noteme
// 		t.MaxCount = maxcount
// 		t.LastPB = lastpb
// 		t.Count = count
// 		t.Type = ty
// 	}

// 	if err == nil {
// 		*start = line.Stop
// 		*until = *start
// 		*mode = "fence"
// 	}

// 	return
// }

// func AstParagraph(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int) (err error) {
// 	//n := node.(*mast.Paragraph)
// 	line := node.Lines().At(0)
// 	//lno := Lineno(line.Start, linenos)

// 	//pnm := node.Kind().String()
// 	if err == nil {
// 		*content += "\n\n"
// 		*start = line.Stop
// 	}
// 	return
// }

// func AstEmphasis(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int) (err error) {
// 	n := node.(*mast.Emphasis)
// 	if err == nil {
// 		*mode = strings.Repeat("*", n.Level)
// 	}
// 	return
// }

// func AstImage(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, linenos []int, alts map[string]*pstructure.Image, dir string) (err error) {
// 	n := node.(*mast.Image)
// 	lineno := Lineno(*start, linenos)
// 	url := string(bytes.TrimSpace(n.Destination))
// 	if len(url) == 0 {
// 		err = perror.Error("image-url", lineno, "missing file]")
// 		return
// 	}
// 	btext := n.Text(blob)
// 	alt := strings.TrimSpace(string(btext))
// 	rest := string(blob[*start:])
// 	k := strings.Index(rest, url)
// 	if k != -1 {
// 		lineno = Lineno(k+*start, linenos)
// 	}
// 	title := string(bytes.TrimSpace(n.Title))
// 	if title != "" {
// 		title = url + " " + title
// 	} else {
// 		title = url
// 	}

// 	cpright := LastCopyRight(doc)
// 	err = NewImage(title, lineno, cpright, doc, dir, alt, alts)
// 	return
// }

// func AstText(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int, alts map[string]*pstructure.Image, dir string) (err error) {
// 	//pnm := node.Kind().String()
// 	n := node.(*mast.Text)
// 	segment := n.Segment
// 	if *mode == "heading" && segment.Start < *until {
// 		*mode = ""
// 		return
// 	}
// 	if *mode == "fence" && segment.Start < *until {
// 		*mode = ""
// 		return
// 	}

// 	lno := Lineno(segment.Start, linenos)
// 	btext := n.Text(blob)
// 	text := string(btext)
// 	m := ptools.MetaChars(text)
// 	if m != "" && !strings.HasPrefix(*mode, "`") {
// 		st := *until
// 		end := segment.Stop
// 		if st > end {
// 			end = len(blob)
// 		}
// 		rest := string(blob[st:end])
// 		k := strings.IndexAny(rest, m)
// 		if strings.TrimSpace(rest[:k]) == "" {
// 			if k != -1 {
// 				count := strings.Count(rest[:k], "\n")
// 				if blob[st] == 10 {
// 					count--
// 				}
// 				lno += count
// 			}
// 		}
// 		if text == "*" || text == "**" {
// 			err = perror.Error("text-meta-unescaped1", lno, "text contains unescaped metachars, probably superfluous whitespace: "+m+"["+text+"]")
// 			return
// 		}
// 		err = perror.Error("text-meta-unescaped2", lno, "text contains unescaped metachars: "+m)
// 		return
// 	}
// 	if strings.HasPrefix(*mode, "*") || strings.HasPrefix(*mode, "`") {
// 		text = *mode + text + *mode
// 		*mode = ""
// 	}
// 	if strings.TrimSpace(text) != "" {
// 		c := doc.LastChapter()
// 		if c == nil {
// 			err = perror.Error("doc-chapter-missing", lno, "no chapter defined yet")
// 			return
// 		}
// 		t := c.LastTopic()
// 		if t == nil {
// 			err = perror.Error("doc-topic-missing", lno, "no topic defined yet")
// 			return
// 		}
// 	}
// 	if strings.Contains(text, ".jpg") {

// 		lno += 2
// 		title := text
// 		err = NewImage(title, lno, "", doc, dir, "", alts)
// 		if err != nil {
// 			return
// 		}
// 	}
// 	if err == nil {
// 		*start = segment.Stop
// 		if n.SoftLineBreak() {
// 			*content += bstrings.RightTrimSpace(text) + "\n"
// 		} else {
// 			*content += text
// 		}
// 	}
// 	return
// }

// func AstCodeSpan(doc *pstructure.Document, node mast.Node, blob []byte, pn mast.Node, start *int, content *string, mode *string, until *int, linenos []int) (code string, err error) {
// 	n := node.(*mast.CodeSpan)
// 	btext := n.Text(blob)
// 	//pnm := node.Kind().String()
// 	code = string(btext)
// 	lineno := Lineno(*start, linenos)
// 	if strings.Contains(code, "\n") {
// 		err = perror.Error("codespan-multiple", lineno, "multiple lines in code")
// 		return
// 	}
// 	if strings.TrimSpace(code) != code {
// 		err = perror.Error("codespan-whitespace", lineno, "whitespace around code is not allowed")
// 		return
// 	}
// 	if code == "" {
// 		err = perror.Error("codespan-empty", lineno, "empty code")
// 		return
// 	}
// 	if err == nil {
// 		*mode = "`"
// 	}

// 	return
// }

// func IsUTF8(reader io.Reader, latin1 bool) (blob []byte, linenos []int, err error) {
// 	buf := bufio.NewReader(reader)
// 	var bbuf bytes.Buffer
// 	cr := byte('\n')
// 	repl := rune(65533)
// 	sum := 0
// 	for {
// 		b, err := buf.ReadBytes(cr)
// 		if err != nil && err != io.EOF {
// 			err = perror.Error("source-read", 0, err)
// 			return nil, nil, err
// 		}
// 		if !utf8.Valid(b) {
// 			err = perror.Error("utf8-noutf8", len(linenos), "line is not valid UTF-8")
// 			return nil, nil, err
// 		}
// 		if bytes.ContainsRune(b, repl) {
// 			err = perror.Error("utf8-repl", len(linenos), "replacement character in line")
// 			return nil, nil, err
// 		}
// 		sum += len(b)
// 		linenos = append(linenos, sum)
// 		if !latin1 {
// 			bbuf.Write(b)
// 			continue
// 		}
// 		bs := ptools.Normalize(string(b), true)
// 		bbuf.WriteString(bs)
// 		bbuf.WriteByte(10)
// 		if err != nil {
// 			break
// 		}
// 	}
// 	return bbuf.Bytes(), linenos, nil
// }

// func Lineno(start int, linenos []int) int {
// 	start++
// 	for i, c := range linenos {
// 		if start <= c {
// 			return i + 1
// 		}
// 	}
// 	return len(linenos) + 1
// }

// func Content(n mast.Node, blob []byte) []byte {
// 	buf := bytes.Buffer{}
// 	for i := 0; i < n.Lines().Len(); i++ {
// 		line := n.Lines().At(i)
// 		buf.Write(line.Value(blob))
// 	}
// 	return buf.Bytes()
// }
