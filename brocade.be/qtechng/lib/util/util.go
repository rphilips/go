package util

import (
	"bufio"
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	qfnmatch "brocade.be/base/fnmatch"
	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

// EMatch extende match
func EMatch(pattern string, qpath string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if qpath == pattern {
		return true
	}
	if qfnmatch.Match(pattern, qpath) {
		return true
	}
	k := strings.IndexAny(pattern, "*[?")
	if k != -1 {
		return false
	}
	pattern += "/*"
	return qfnmatch.Match(pattern, qpath)
}

// About changes the comment form to '//'
func About(blob []byte) (result []byte) {
	buffer := bytes.NewBuffer(blob)
	eol := byte('\n')
	slash := []byte("//")
	delim := ""
	ok := -1
	stop := false
	for {
		if stop {
			if ok == 1 {
				return
			}
			return blob[:]
		}
		if ok == 1 {
			line := buffer.Bytes()
			result = append(result, line...)
			return
		}
		line, err := buffer.ReadBytes(eol)
		if err != nil && err != io.EOF {
			return blob[:]
		}
		if err == io.EOF {
			stop = true
		}
		s := string(line)
		if ok == -1 {
			s = strings.TrimSpace(s)
			if strings.HasPrefix(s, `"""`) || strings.HasPrefix(s, `'''`) {
				delim = s[:3]
				bdelim := []byte(delim)
				k := bytes.Index(line, bdelim)
				line = line[k+3:]
				s = s[3:]
				ok = 0
			}
			if delim == "" {
				result = append(result, line...)
				if s != "" {
					ok = 1
				}
				continue
			}
			if strings.HasSuffix(s, delim) {
				bdelim := []byte(delim)
				k := bytes.LastIndex(line, bdelim)
				line = append(line[:k], line[k+3:]...)
				ok = 1
			}
			result = append(result, slash...)
			result = append(result, line...)
			continue
		}
		if ok == 0 {
			s = strings.TrimRightFunc(s, unicode.IsSpace)
			if strings.HasSuffix(s, delim) {
				bdelim := []byte(delim)
				k := bytes.LastIndex(line, bdelim)

				if k > 0 {
					line = append(line[:k], line[k+3:]...)
				} else {
					line = line[3:]
				}
				ok = 1
			}
			result = append(result, slash...)
			result = append(result, line...)
			continue
		}
	}
}

// Time make a string
func Time(times ...string) string {
	t := ""
	if len(times) > 0 {
		t = times[0]
	}
	if t == "" {
		h := time.Now()
		t = h.Format(time.RFC3339)
	}
	parts := regexp.MustCompile("[^1234567890]+").Split(t, -1)
	if len(parts) < 3 {
		return t
	}
	year, _ := strconv.Atoi(parts[0])
	month, _ := strconv.Atoi(parts[1])
	day, _ := strconv.Atoi(parts[2])
	hour := 0
	if len(parts) > 3 {
		hour, _ = strconv.Atoi(parts[3])
	}
	min := 0
	if len(parts) > 4 {
		min, _ = strconv.Atoi(parts[4])
	}
	sec := 0
	if len(parts) > 5 {
		sec, _ = strconv.Atoi(parts[5])
	}
	t = time.Date(year, time.Month(month), day, hour, min, sec, 0, time.UTC).Format(time.RFC3339)
	t = t[:19]
	return t
}

// Comment maakt een commentaar lijn uit een slice of strings
func Comment(cmt interface{}) string {
	c := make([]string, 0)
	if cmt == nil {
		return ""
	}
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

// NoUTF8 zoekt naar lijnen die geen geldige UTF-8 bevatten: geeft een slice van [row, col] terug (en een errorr)
func NoUTF8(reader io.Reader) (body []byte, result [][2]int, err error) {
	repl := rune(65533)
	body, err = ioutil.ReadAll(reader)
	result = make([][2]int, 0)
	if err != nil {
		return
	}
	if utf8.Valid(body) && !bytes.ContainsRune(body, repl) {
		return
	}
	lines := strings.Split(string(body), "\n")
	for count, line := range lines {
		if utf8.ValidString(line) && !strings.ContainsRune(line, repl) {
			continue
		}
		good := strings.ToValidUTF8(line, "\n")
		good = strings.ReplaceAll(good, string(repl), "\n")
		parts := strings.Split(string(good), "\n")

		total := ""
		for c, part := range parts {
			if c == len(parts) {
				break
			}
			total += part + "\n"
			result = append(result, [2]int{count + 1, len([]rune(total))})
		}
	}
	return
}

// Info creates information lines from a string
func Info(s string, prefix string) string {
	lines := strings.SplitN(s, "\n", -1)
	result := make([]string, 0)
	cutset := "\n\r \t"
	k := -1
	before := ""
	//Log("prefix:", prefix, "\ns: ["+s+"]")
	for _, line := range lines {
		line = strings.TrimRight(line, cutset)
		if line == "" {
			continue
		}
		if before == "" {
			k = strings.Index(line, ":")
			if k == -1 {
				continue
			}
			k++
			rest := line[k:]
			x := strings.TrimLeft(rest, cutset)
			k += len(rest) - len(x)
			if x != "" {
				result = append(result, x)
			}
			before = strings.Repeat(" ", k)
			continue
		}
		if strings.HasPrefix(line, before) {
			result = append(result, line[k:])
			continue
		}
		line = strings.TrimLeft(line, cutset)
		result = append(result, line)
		continue
	}
	return strings.Join(result, "\n")
}

// Fix removes superfluous characters
func Fix(s string) string {
	x := strings.TrimSpace(s)
	if strings.HasPrefix(x, "«") && strings.HasSuffix(x, "»") {
		return strings.TrimSuffix(strings.TrimPrefix(x, "«"), "»")
	}
	if strings.HasPrefix(x, "⟦") && strings.HasSuffix(x, "⟧") {
		return strings.TrimSuffix(strings.TrimPrefix(x, "⟦"), "⟧")
	}
	return x
}

// Digest calculates the 56 first hexcodes of SHA-512
func Digest(blob []byte) string {
	sum := sha512.Sum512(blob)
	return hex.EncodeToString(sum[:28])
}

// Canon cleans paths
func Canon(s string) string {
	s = strings.ReplaceAll(s, "\\", "/")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.TrimRight(s, "/")
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	return s
}

// MakeBytes maakt een []byte van een stuk data
func MakeBytes(data interface{}) (b []byte, err error) {
	switch v := data.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case *string:
		return []byte(*v), nil
	case bytes.Buffer:
		b, err = ioutil.ReadAll(&v)
		return
	case *bytes.Buffer:
		b, err = ioutil.ReadAll(v)
		return
	case io.Reader:
		b, err = ioutil.ReadAll(v)
		return
	default:
		b, err := json.MarshalIndent(data, "", "    ")
		return b, err
	}
}

// ObjectSplitter maakt een [][]byte van een blob. Elke oneven index is een kandidaat object
func ObjectSplitter(blob []byte) (result [][]byte) {
	result = append(make([][]byte, 0), make([]byte, 0))
	max := len(result) - 1
	rest := blob
	for {
		if len(rest) == 0 {
			break
		}
		k := bytes.Index(rest, []byte("4_"))
		if k < 0 {
			result[max] = append(result[max], rest...)
			break
		}
		if k == 0 {
			result[max] = append(result[max], byte('4'), byte('_'))
			rest = rest[2:]
			continue
		}
		if k+2 == len(rest) {
			result[max] = append(result[max], rest...)
			break
		}
		if !IsObjStarter(rest[k-1:]) {
			result[max] = append(result[max], rest[:k+2]...)
			rest = rest[k+2:]
			continue
		}
		result[max] = append(result[max], rest[:k-1]...)
		rest = rest[k-1:]
		prev := rune(rest[0])
		obj := ""
		switch prev {
		case 'm', 'i', 't':
			obj = mgobble(rest[3:])
		case 'r':
			obj = rgobble(rest[3:])
		case 'l':
			obj = lgobble(rest[3:])
		}
		result = append(result, rest[:3+len(obj)])
		rest = rest[3+len(obj):]
		result = append(result, []byte{})
		max = len(result) - 1
	}
	return
}

func mgobble(rest []byte) (obj string) {
	if len(rest) == 0 {
		return ""
	}
	next := rest[0]
	if strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", next) == -1 {
		return ""
	}
	bobj := make([]byte, 0)
	for _, b := range rest {
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", b) == -1 {
			break
		}
		bobj = append(bobj, b)
		continue
	}
	obj = string(bobj)
	if !strings.HasSuffix(obj, "4") {
		return obj
	}
	if IsObjStarter(rest[len(bobj)-2:]) {
		return string(bobj[:len(bobj)-2])
	}
	return obj
}

func rgobble(rest []byte) (obj string) {
	if len(rest) == 0 {
		return ""
	}
	next := rest[0]
	if strings.IndexByte("abcdefghijklmnopqrstuvwxyz", next) == -1 {
		return ""
	}
	bobj := make([]byte, 0)
	for _, b := range rest {
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyz0123456789_", b) == -1 {
			break
		}
		bobj = append(bobj, b)
		continue
	}
	obj = string(bobj)
	k := strings.Index(obj, "--")
	if k != -1 {
		obj = obj[:k]
	}
	k = strings.Index(obj, "4_")
	if k == -1 {
		return obj
	}
	if IsObjStarter(rest[k-1:]) {
		return obj[:k-1]
	}
	return obj
}

func lgobble(rest []byte) (obj string) {
	if len(rest) == 0 {
		return ""
	}
	r := rest
	bobj := make([]byte, 0)
	for _, b := range r {
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ_", b) == -1 {
			break
		}
		bobj = append(bobj, b)
	}
	obj = string(bobj)
	if obj == "" {
		return ""
	}
	k := strings.Count(obj, "_")
	if k > 1 {
		parts := strings.SplitN(obj, "_", -1)
		obj = parts[0] + "_" + parts[1]
		k = 1
	}

	if k == 1 {
		parts := strings.SplitN(obj, "_", -1)
		if parts[0] == "" {
			return ""
		}
		algo := parts[0][1:]
		if strings.IndexRune("NEDFU", rune(obj[0])) == -1 {
			return ""
		}
		if algo != "" && algo != "js" && algo != "py" && algo != "php" {
			return ""
		}
	}

	return obj
}

// IsObjStarter keert met true terug als de byteslice kan beginnen met een object
func IsObjStarter(rest []byte) bool {
	if len(rest) < 4 {
		return false
	}
	if rest[1] != byte('4') {
		return false
	}
	if rest[2] != byte('_') {
		return false
	}
	if strings.IndexByte("milrt", rest[0]) == -1 {
		return false
	}
	mode := rest[0]
	switch mode {
	case byte('m'), byte('i'):
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", rest[3]) == -1 {
			return false
		}
		return !IsObjStarter(rest[3:])
	case byte('r'):
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyz", rest[3]) == -1 {
			return false
		}
		return !IsObjStarter(rest[3:])
	case byte('l'):
		if len(rest) < 4 {
			return false
		}
		if strings.IndexByte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", rest[3]) == -1 {
			return false
		}
		return !IsObjStarter(rest[4:])
	default:
		return false
	}
}

// ExtractLineno extracts the line form a body
func ExtractLineno(msg string, body []byte) (int, string) {
	parts := strings.SplitN(msg, ":", 3)
	if len(parts) < 2 {
		return -1, ""
	}
	lineno, err := strconv.Atoi(parts[1])
	if err != nil {
		return -1, ""
	}
	line := ""
	if lineno > 0 {
		dlm := byte('\n')
		r := bufio.NewReader(bytes.NewReader(body))
		for i := 0; i < lineno; i++ {
			xine, e := r.ReadSlice(dlm)
			if i == lineno-1 {
				line = string(xine)
				break
			}
			if e != nil {
				break
			}
		}
	}
	return lineno, line
}

// ExtractMsg removers bolierplate from PEG parser
func ExtractMsg(msg, fname string) string {
	msg = strings.ReplaceAll(msg, fname, "")
	return strings.TrimLeft(msg, ": ")
}

// LStrip a string
func LStrip(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

// RStrip a string
func RStrip(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// QPartition splits a qpath
func QPartition(qpath string) (dir string, base string) {
	if qpath == "" {
		return "/", ""
	}
	k := strings.LastIndex(qpath, "/")
	switch {
	case k < 0:
		return "", qpath
	case k == 0:
		return "/", qpath[1:]
	case k == len(qpath)-1:
		return qpath, qpath[:k-1]
	default:
		return qpath[:k], qpath[k+1:]
	}
}

// Log logs bug information
func Log(v ...interface{}) {
	filename := qregistry.Registry["qtechng-log"]
	if filename == "" {
		return
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, v...)
	fmt.Fprintln(f, "===")
}

// GetPy find the suitable python executable
func GetPy(pyscript string) string {
	cwd, e := os.Getwd()
	if e != nil {
		return ""
	}
	if !filepath.IsAbs(pyscript) {
		pyscript, _ = qfs.AbsPath(path.Join(cwd, pyscript))
	}
	reader, err := os.Open(pyscript)
	if err != nil {
		return ""
	}
	read := bufio.NewReader(reader)
	first, _ := read.ReadString('\n')
	reader.Close()
	pyfound := ""
	k := strings.Index(first, "#")
	if k != -1 {
		first := first[k:]
		k = strings.Index(first, "py2")
		if k == -1 {
			k = strings.Index(first, "python2")
		}
		if k != -1 {
			pyfound = "py2"
		}
		if k == -1 {
			k = strings.Index(first, "python3")
			if k == -1 {
				k = strings.Index(first, "py3")
			}
			if k != -1 {
				pyfound = "py3"
			}
		}
	}
	if pyfound != "" {
		return pyfound
	}
	subs := strings.SplitN(filepath.ToSlash(pyscript), "/", -1)
	for {
		if len(subs) == 0 {
			return ""
		}
		subs = subs[:len(subs)-1]
		fname := path.Join(subs...)
		fname = path.Join(fname, "brocade.json")
		blob, e := qfs.Fetch("brocade.json")
		if e != nil {
			continue
		}
		cfg := make(map[string]interface{})
		e = json.Unmarshal(blob, &cfg)
		if e != nil {
			continue
		}
		v, ok := cfg["py3"]
		if !ok {
			return "py2"
		}
		if v.(bool) {
			return "py3"
		}
		return "py2"
	}
}

// BuildArgs searche for arguments for macro's or x4_
func BuildArgs(s string) (args []string, until string, msg string) {

	if !strings.HasPrefix(s, "(") {
		msg = ")"
		args = nil
		until = ""
		return
	}

	until = "("
	args = make([]string, 0)
	closer := ' '
	arg := ""

	nest := 0
	msg = "no end `)`"
	stop := false

	for _, ch := range s[1:] {
		if stop {
			break
		}
		until += string(ch)

		if closer != ' ' {
			arg += string(ch)
			if closer == ch {
				closer = ' '
			}
			continue
		}
		switch ch {
		case ')':
			if nest == 0 {
				msg = ""
				stop = true
				args = append(args, strings.TrimSpace(arg))
				continue
			}
			arg += string(ch)
			nest--
		case '(':
			arg += string(ch)
			nest++

		case '«':
			arg += string(ch)
			closer = '»'

		case '⟦':
			arg += string(ch)
			closer = '⟧'

		case '"':
			arg += string(ch)
			closer = '"'

		case ',':
			args = append(args, strings.TrimSpace(arg))
			arg = ""
		default:
			arg += string(ch)
		}
	}
	if len(args) == 1 && len(args[0]) == 0 {
		args = []string{}
	}

	if closer != ' ' {
		msg = "No `" + string(closer) + "` found"
	}
	return args, until, msg

}

// CleanArg removes uperfluous whitespace, « and ⟦
func CleanArg(s string) string {
	s = strings.TrimSpace(s)

	if s == "" {
		return s
	}

	if strings.HasPrefix(s, "«") {
		return strings.TrimSuffix(s, "»")
	}

	if strings.HasPrefix(s, "⟦") {
		return strings.TrimSuffix(s, "⟧")
	}
	return s
}

// BlobSplit splits a []byte according to a regexp
func BlobSplit(blob []byte, split []string, qreg bool) [][]byte {

	parts := make([]string, len(split))

	for i, s := range split {
		part := s
		if !qreg {
			part = "(^|\\n)" + regexp.QuoteMeta(s) + "\\s+"
		}
		parts[i] = part
	}
	re := regexp.MustCompile(strings.Join(parts, "|"))
	find := re.FindAllIndex(blob, -1)
	if len(find) == 0 {
		return [][]byte{blob}
	}
	result := make([][]byte, len(find)+1)

	result[0] = blob[:find[0][0]]

	for i := 0; i < len(find); i++ {
		if i < len(find)-1 {
			result[i+1] = blob[find[i][0]+1 : find[i+1][0]]
		} else {
			result[i+1] = blob[find[i][0]+1:]
		}
	}
	return result
}

// Decomment haalt beginnede  '//' commentaar weg
func Decomment(blob []byte, cmts string) (buf *bytes.Buffer) {
	buf = new(bytes.Buffer)
	content := About(blob)
	lines := bytes.SplitN(content, []byte("\n"), -1)
	preamble := true

	for _, line := range lines {
		if preamble && len(line) > 1 && line[0] == byte('/') && line[1] == byte('/') {
			buf.Write(line)
			buf.WriteRune('\n')
			continue
		}
		preamble = false

		if len(line) < 2 {
			buf.Write(line)
			buf.WriteRune('\n')
			continue
		}

		xline := bytes.TrimLeft(line, " \t")
		if len(xline) < 2 {
			buf.Write(line)
			buf.WriteRune('\n')
			continue
		}

		if line[0] == byte('/') && line[1] == byte('/') {
			continue
		}
		if cmts != "" {
			line, _ = mdecomment(line, cmts)
		}
		buf.Write(line)
		buf.WriteRune('\n')
		continue
	}
	return
}

func mdecomment(line []byte, cmts string) ([]byte, []byte) {
	k := bytes.IndexAny(line, cmts)
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
		x, y := mdecomment(line[1:], cmts)
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
		x, y := mdecomment(line[k+1:], cmts)
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

	x, y := mdecomment(line[f+1:], cmts)
	return append(line[:f+1], x...), y

}

// Embrace creates a delimited string
func Embrace(s string) string {
	if s == "" {
		return ""
	}
	rs := []rune(s)
	delim := false
	if unicode.IsSpace(rs[0]) {
		delim = true
	}
	if !delim && len(rs) > 1 && unicode.IsSpace(rs[len(rs)-1]) {
		delim = true
	}
	k := strings.IndexAny(s, "«»⟦⟧\n")
	if k < 0 {
		if !delim {
			return s
		}
		return "«" + s + "»"
	}
	k = strings.IndexAny(s, "«»⟦⟧")
	if k < 0 {
		return "«" + s + "»"
	}
	k = strings.IndexAny(s, "«»")

	if k < 0 {
		return "«" + s + "»"
	}
	return "⟦" + s + "⟧"
}

// Ignore ignores part of the string
func Ignore(s []byte) []byte {
	if len(s) == 0 {
		return s
	}
	k := bytes.Index(s, []byte("<ignore>"))
	if k == -1 {
		return s
	}
	rest := s[k:]
	l := bytes.Index(rest, []byte("</ignore>"))
	if l == -1 {
		return s[:k]
	}
	return Ignore(append(s[:k], s[k+l+9:]...))
}
