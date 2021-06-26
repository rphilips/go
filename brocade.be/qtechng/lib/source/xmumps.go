package source

import (
	"bytes"
	"strings"

	qmumps "brocade.be/base/mumps"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qutil "brocade.be/qtechng/lib/util"
)

// WidgetsListToMumps bereidt een verzameling van Lgcodes
func WidgetsListToMumps(batchid string, widgets []*qofile.Widget, buf *bytes.Buffer) (errs []error) {
	for _, pwidget := range widgets {
		mumps, err := pwidget.Mumps(batchid)
		if err != nil {
			errs = append(errs, err)
		}
		qmumps.Println(buf, mumps)
	}
	if len(errs) == 0 {
		return nil
	}
	return errs
}

// XFileToMumps schrijft een M file naar een buffer
func (xfile *Source) XFileToMumps(batchid string, buf *bytes.Buffer) error {
	content, err := xfile.Fetch()
	if err != nil {
		return err
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
	if len(content) != 0 {
		err = qobject.Loads(xf, content, true)
		if err != nil {
			return err
		}
	}
	objectlist := xf.Objects()
	textmap := make(map[string]string)

	for _, obj := range objectlist {
		ty := obj.Type()
		if ty == "text" {
			name := strings.TrimSpace(strings.TrimPrefix(obj.Name(), "text"))
			textmap[name] = obj.(*qofile.Widget).Body
		}
	}
	env := xfile.Env()
	notreplace := xfile.NotReplace()
	objectmap := make(map[string]qobject.Object)
	bufmac := new(bytes.Buffer)
	_, err = ResolveText(env, content, "trilm", notreplace, objectmap, textmap, bufmac, "", xfile.String())
	if err != nil {
		return err
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
	err = qobject.Loads(xf, content, false)
	if err != nil {
		return err
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
	errs := WidgetsListToMumps(batchid, widgets, buf)
	if errs == nil {
		return nil
	}

	return qerror.ErrorSlice(errs)
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
