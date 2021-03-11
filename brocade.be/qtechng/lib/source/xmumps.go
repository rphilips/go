package source

import (
	"bytes"
	"strings"

	qmumps "brocade.be/base/mumps"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// WidgetsListToMumps bereidt een verzameling van Lgcodes
func WidgetsListToMumps(batchid string, widgets []*qofile.Widget, buf *bytes.Buffer) {
	for _, pwidget := range widgets {
		mumps := pwidget.Mumps(batchid)
		qmumps.Println(buf, mumps)
	}
	return
}

// XFileToMumps schrijft een M file naar een buffer
func (xfile *Source) XFileToMumps(batchid string, buf *bytes.Buffer) {
	content, err := xfile.Fetch()
	if err != nil {
		return
	}
	content = qutil.Decomment(content).Bytes()
	bufnoc := new(bytes.Buffer)
	//qpath := xfile.String()

	content = qutil.About(content)
	lines := bytes.SplitN(content, []byte("\n"), -1)
	preamble := true

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if line[0] == byte('/') {
			if preamble {
				continue
			}
		}
		preamble = false
		line, _ := xdecomment(line)

		if len(line) == 0 {
			continue
		}

		bufnoc.Write(line)
		bufnoc.WriteRune('\n')
	}
	content = bufnoc.Bytes()
	xf := new(qofile.XFile)
	xf.SetEditFile(xfile.String())
	xf.SetRelease(xfile.Release().String())
	err = qobject.Loads(xf, content)
	if err != nil {
		//fmt.Println("err:", err.Error())
		return
	}
	objectlist := xf.Objects()
	textmap := make(map[string]string)

	for _, obj := range objectlist {
		ty := obj.Type()
		if ty == "text" {
			id := obj.String()
			name := strings.SplitN(id, " ", 1)[0]
			textmap[name] = obj.(*qofile.Widget).Body
		}
	}
	env := xfile.Env()
	notreplace := xfile.NotReplace()
	objectmap := make(map[string]qobject.Object)
	bufmac := new(bytes.Buffer)
	_, err = ResolveText(env, content, "trilm", notreplace, objectmap, textmap, bufmac, "")
	if err != nil {
		//fmt.Println("err:", err.Error())
	}
	content = bufmac.Bytes()

	lines = bytes.SplitN(content, []byte("\n"), -1)
	buffer := new(bytes.Buffer)
	for _, line := range lines {
		if len(line) == 0 {
			buffer.WriteRune('\n')
			continue
		}
		line, _ := xdecomment(line)
		if len(line) != 0 {
			buffer.Write(line)
		}
		buffer.WriteRune('\n')
	}
	content = buffer.Bytes()
	err = qobject.Loads(xf, content)
	if err != nil {
		return
	}
	objectlist = xf.Objects()
	widgets := make([]*qofile.Widget, 0)
	for _, obj := range objectlist {
		ty := obj.Type()
		if ty == "text" {
			continue
		}
		widgets = append(widgets, obj.(*qofile.Widget))
	}
	WidgetsListToMumps(batchid, widgets, buf)
	return
}

func xsplit(lines [][]byte) map[string][]byte {
	objs := make(map[string][]byte)
	//name := ""

	return objs
}

func xdecomment(line []byte) ([]byte, []byte) {
	k := bytes.Index(line, []byte("//"))
	if k == -1 {
		return line, []byte{}
	}
	if k == 0 {
		return []byte{}, line
	}

	pre := line[:k]
	l := bytes.IndexAny(pre, `"«⟦`)
	if l == -1 {
		if line[k-1] != byte(':') {
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

	x, y := xdecomment(line[f+1:])
	return append(line[:f+1], x...), y
}

func xtransform(line []byte) []byte {
	return line
}
