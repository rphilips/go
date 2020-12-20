package source

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// MFileToMumps schrijft een M file naar een buffer
func (mfile *Source) MFileToMumps(batchid string, buf *bytes.Buffer) {
	content, err := mfile.Fetch()
	if err != nil {
		return
	}
	bufnoc := new(bytes.Buffer)
	qpath := mfile.String()
	version := mfile.Release().String()
	_, base := qutil.QPartition(qpath)
	ext := path.Ext(base)
	root := strings.TrimSuffix(base, ext)
	bufnoc.WriteString(root + " ;" + batchid + "\n")
	bufnoc.WriteString(" ; version=" + version + "\n")

	meta, err := qmeta.Meta{}.New(version, qpath)
	if err == nil {
		bufnoc.WriteString(" ; cuser=" + meta.Cu + "\n")
		bufnoc.WriteString(" ; ctime=" + meta.Ct + "\n")
		bufnoc.WriteString(" ; muser=" + meta.Mu + "\n")
		bufnoc.WriteString(" ; mtime=" + meta.Mt + "\n")
	}

	content = qutil.About(content)
	lines := bytes.SplitN(content, []byte("\n"), -1)
	preamble := true

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == byte('/') {
			if preamble {
				bufnoc.WriteString(" ; ")
				bufnoc.Write(line)
				bufnoc.WriteRune('\n')
				continue
			}
		}
		if preamble {
			bufnoc.WriteString("ltechbeg ; path=" + qpath + "\n")
			preamble = false
		}
		line, comment := mdecomment(line)
		if len(comment) != 0 && comment[0] == byte('/') {
			comment = []byte{}
		}

		if len(line) == 0 && len(comment) == 0 {
			continue
		}

		if len(line) == 0 {
			bufnoc.WriteRune(' ')
			bufnoc.Write(comment)
			bufnoc.WriteRune('\n')
			continue
		}
		bufnoc.Write(line)
		if len(comment) != 0 {
			bufnoc.WriteString("  ")
			bufnoc.Write(comment)
		}
		bufnoc.WriteRune('\n')
	}
	content = bufnoc.Bytes()
	env := mfile.Env()
	notreplace := mfile.NotReplace()
	objectmap := make(map[string]qobject.Object)
	bufmac := new(bytes.Buffer)
	_, err = ResolveText(env, content, "rilm", notreplace, objectmap, nil, bufmac, "")
	if err != nil {
		fmt.Println("err:", err.Error())
	}
	content = bufmac.Bytes()

	lines = bytes.SplitN(content, []byte("\n"), -1)

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		line, comment := mdecomment(line)
		if len(comment) != 0 && comment[0] == byte('/') {
			comment = []byte{}
		}
		if len(line) == 0 && len(comment) == 0 {
			continue
		}
		if len(line) == 0 {
			buf.WriteRune(' ')
			buf.Write(comment)
			buf.WriteRune('\n')
			continue
		}
		xline := mtransform(line, comment)
		if len(xline) != 0 {
			buf.Write(xline)
			buf.WriteRune('\n')
		}
	}
	buf.WriteString("ltechend ; path=" + qpath + "\n")
}

func mdetag(line []byte) ([]byte, []byte, string) {
	short := bytes.TrimLeft(line, "%1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	k := len(line) - len(short)
	if k == 0 {
		return []byte{}, bytes.TrimSpace(line), ""
	}
	if len(line) == k {
		return line, []byte{}, ""
	}
	fun := ""
	if line[k] == byte(' ') {
		stag := string(line[:k])
		if stag == "def" || stag == "sub" || stag == "fn" {
			fun = stag
			line = bytes.TrimSpace(line[k+1:])
			short = bytes.TrimLeft(line, "%1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
			k = len(line) - len(short)
			if k == 0 {
				return []byte{}, bytes.TrimSpace(line), fun
			}
			if len(line) == k {
				return line, []byte{}, fun
			}
		}
	}
	name := line[:k]
	short = bytes.TrimSpace(short)
	if len(short) == 0 {
		return name, short, fun
	}
	if short[0] != byte('(') {
		if fun != "" {
			short = bytes.TrimLeft(short, ": \t")
		}
		return name, short, fun
	}
	pargs := bytes.TrimLeft(short[1:], "\t ,%1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	if len(pargs) == 0 || pargs[0] != byte(')') {
		return name, short, fun
	}
	k = len(short) - len(pargs) + 2

	name = append(name, bytes.ReplaceAll(short[:k], []byte(" "), []byte{})...)

	if fun != "" {
		short = bytes.TrimLeft(short[k:], ": \t")
	}
	return name, short, fun
}

func mtransform(line []byte, comment []byte) []byte {
	if len(line) == 0 {
		if len(comment) == 0 {
			return line
		}
		return append([]byte(" "), comment...)
	}

	tag, line, fun := mdetag(line)

	if len(line) == 0 {
		if len(tag) == 0 && len(comment) == 0 {
			return line
		}
		buffer := new(bytes.Buffer)
		if len(tag) != 0 {
			buffer.Write(tag)
		}
		buffer.WriteString(" ;")
		if len(comment) == 0 {
			buffer.WriteString(fun)
		} else {
			buffer.Write(comment[1:])
		}
		return buffer.Bytes()
	}

	dots := []byte{}
	if rune(line[0]) == '.' {
		xline := bytes.TrimLeft(line, ". \t")
		dots = line[:len(line)-len(xline)+1]
		k := bytes.Count(dots, []byte("."))
		dots = bytes.Repeat([]byte("."), k)
		line = xline
	}

	result, winarg := mbeautify(string(line))

	buffer := new(bytes.Buffer)

	if len(result) == 0 {
		if len(tag) == 0 && len(comment) == 0 {
			return []byte{}
		}
		buffer := new(bytes.Buffer)
		if len(tag) != 0 {
			buffer.Write(tag)
		}
		buffer.WriteString(" ;")
		if len(comment) == 0 {
			buffer.WriteString(fun)
		} else {
			buffer.Write(comment[1:])
		}
		return buffer.Bytes()
	}

	if len(tag) != 0 {
		buffer.Write(tag)
	}
	buffer.WriteRune(' ')
	if len(dots) != 0 {
		buffer.Write(dots)
		buffer.WriteRune(' ')
	}

	buffer.WriteString(result)

	if len(comment) != 0 {
		if !winarg {
			buffer.WriteRune(' ')
		}
		buffer.WriteRune(' ')
		buffer.Write(comment)
	}

	return buffer.Bytes()
}

func mbeautify(line string) (string, bool) {
	buffer := new(bytes.Buffer)
	instring := false
	inarg := false
	winarg := false
	waitcmd := false
	witharg := false
	for _, ru := range line {
		if ru == '"' {
			instring = !instring
		}
		if waitcmd && !instring && (ru == ' ' || ru == '\t') {
			continue
		}
		if instring || (ru != ' ' && ru != '\t') {
			if waitcmd {
				if !winarg {
					buffer.WriteRune(' ')
					winarg = true
				}
				buffer.WriteRune(' ')
				buffer.WriteRune(ru)
				waitcmd = false
				witharg = false
				continue
			}
			if inarg && !winarg {
				buffer.WriteRune(' ')
				buffer.WriteRune(ru)
				winarg = true
				witharg = true
				continue
			}
			buffer.WriteRune(ru)
			continue
		}
		if inarg {
			waitcmd = true
			inarg = false
			continue
		}
		inarg = true
		winarg = false
		waitcmd = false
	}

	return buffer.String(), witharg
}

func mdecomment(line []byte) ([]byte, []byte) {
	k := bytes.IndexAny(line, "/;")
	if k == -1 {
		return line, []byte{}
	}
	if k == 0 {
		if line[k] == byte(';') {
			return []byte{}, line
		}
		if len(line) == k+1 {
			return line, []byte{}
		}
		if line[k+1] == byte('/') {
			return []byte{}, line
		}
		x, y := mdecomment(line[1:])
		return line[:1+len(x)], y
	}

	pre := line[:k]
	l := bytes.IndexAny(pre, `"«⟦`)
	if l == -1 {
		if line[k] == byte(';') {
			return line[:k], line[k:]
		}
		if len(line) == k+1 {
			return line, []byte{}
		}
		if line[k+1] == byte('/') {
			return line[:k], line[k:]
		}
		x, y := mdecomment(line[k+1:])
		return line[:1+len(x)], y
	}

	t := '"'

	switch {
	case line[l] == byte('"'):
		t = '"'
	case bytes.HasPrefix(line[l:], []byte("«")):
		t = '»'
	default:
		t = '⟧'
	}

	f := bytes.IndexRune(line[l+1:], t)

	if f == -1 {
		return line, []byte{}
	}
	f += l + 1

	x, y := mdecomment(line[f+1:])
	return append(line[:f+1], x...), y

}
