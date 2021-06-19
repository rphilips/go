package mumps

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

type MUMPS []M

type M struct {
	Subs   []string
	Value  string
	Action string
}

func (m M) String() string {
	switch m.Action {
	case "exec":
		return m.Value
	case "kill":
		if len(m.Subs) == 0 {
			return ""
		}
		name := m.Subs[0]
		if len(m.Subs) == 1 {
			return "k " + name
		}
		escaped := make([]string, len(m.Subs)-1)
		for i, s := range m.Subs[1:] {
			escaped[i] = Tom(s)
		}
		prefix := "("
		if strings.HasSuffix(name, ")") {
			name = name[:len(name)-1]
			prefix = ","
		}
		return "k " + name + prefix + strings.Join(escaped, ",") + ")"
	case "set":
		if len(m.Subs) == 0 {
			return ""
		}
		name := m.Subs[0]
		if len(m.Subs) == 1 {
			return "s " + name + "=" + Tom(m.Value)
		}
		escaped := make([]string, len(m.Subs)-1)
		for i, s := range m.Subs[1:] {
			escaped[i] = Tom(s)
		}
		prefix := "("
		if strings.HasSuffix(name, ")") {
			name = name[:len(name)-1]
			prefix = ","
		}
		return "s " + name + prefix + strings.Join(escaped, ",") + ")=" + Tom(m.Value)
	}
	return ""
}

func Tom(s string) string {
	if len(s) == 0 {
		return `""`
	}

	bs := []byte(s)
	var sb strings.Builder
	sb.Grow(len(bs) + 2)
	sb.WriteString(`"`)
	buf := make([]string, 0)
	for _, b := range bs {
		if b > 31 && b < 128 {
			if len(buf) != 0 {
				prefix := strings.Join(buf, ",")
				buf = buf[:0]
				sb.WriteString(`"_$C(`)
				sb.WriteString(prefix)
				sb.WriteString(`)_"`)
			}
			if b == 34 {
				sb.WriteString(`""`)
				continue
			}
			sb.WriteByte(b)
			continue
		}
		buf = append(buf, strconv.Itoa(int(b)))
	}
	if len(buf) != 0 {
		prefix := strings.Join(buf, ",")
		sb.WriteString(`"_$C(`)
		sb.WriteString(prefix)
		sb.WriteString(`)`)
	} else {
		sb.WriteString(`"`)
	}
	return strings.TrimPrefix(sb.String(), `""_`)
}

func Set(mumps MUMPS, subs []string, value string) MUMPS {
	msub := M{
		Subs:   subs[:],
		Value:  value,
		Action: "set",
	}
	x := append(mumps, msub)
	return x
}

func Kill(mumps MUMPS, subs []string) MUMPS {
	msub := M{
		Subs:   subs[:],
		Action: "kill",
	}
	x := append(mumps, msub)
	return x
}

func Exec(mumps MUMPS, statement string) MUMPS {
	msub := M{
		Value:  statement,
		Action: "exec",
	}
	x := append(mumps, msub)
	return x
}

func Println(w io.Writer, mumps MUMPS) {
	for _, m := range mumps {
		fmt.Fprintln(w, m.String())
	}
}

func Print(w io.Writer, mumps MUMPS) {
	for _, m := range mumps {
		fmt.Fprint(w, m.String())
	}
}

// PipeTo writes M instructions to M
func PipeTo(mdb string, buffers []*bytes.Buffer) (err error) {
	cmd, err := newMCMD(mdb)
	if err != nil {
		return
	}
	stdin, e := cmd.StdinPipe()
	if e != nil {
		return e
	}
	go func() {
		defer stdin.Close()
		for _, b := range buffers {
			if b.Len() == 0 {
				continue
			}
			io.Copy(stdin, b)
		}
		io.WriteString(stdin, "\n\nq\nh\n")
	}()
	_, e = cmd.CombinedOutput()
	return e
}

// PipeLineTo writes M instructions to M
func PipeLineTo(mdb string, reader *bufio.Reader) (err error) {
	cmd, err := newMCMD(mdb)
	if err != nil {
		return
	}

	stdin, e := cmd.StdinPipe()
	if e != nil {
		return e
	}
	go func() {
		defer stdin.Close()
		io.Copy(stdin, reader)
		io.WriteString(stdin, "\n\nq\nh\n")
	}()
	_, e = cmd.CombinedOutput()
	return e
}

func newMCMD(mdb string) (cmd *exec.Cmd, err error) {
	rou := qregistry.Registry["m-import-auto-exe"] // m-import-auto-exe  : ["anetcache", "%RunDS^bqtm"]
	//rou := ""
	rouparts := make([]string, 0)
	if rou != "" {
		e := json.Unmarshal([]byte(rou), &rouparts)
		if e != nil {
			return nil, fmt.Errorf("registry value `m-import-auto-exe` is not JSON: `%s`", e.Error())
		}
	} else {
		h := time.Now()
		t := h.Format(time.RFC3339)
		t = strings.ReplaceAll(t, ":", ".")
		t = strings.ReplaceAll(t, "+", ".")
		target := "mumpssinc" + t + ".*.txt"
		target, _ = qfs.TempFile("", target)

		rouparts = []string{
			"qtechng",
			"fs",
			"store",
			target,
		}
	}
	inm := rouparts[0]
	inm, err = exec.LookPath(inm)
	if err != nil {
		return
	}
	if len(rouparts) == 1 {
		cmd = exec.Command(inm)
	} else {
		cmd = exec.Command(inm, rouparts[1:]...)
	}
	if mdb == "" {
		mdb = qregistry.Registry["m-db"]
	}
	cmd.Dir = mdb
	return cmd, nil
}

// Compile tests if a m script compiles:
func Compile(scriptm string) error {
	compiler := qregistry.Registry["m-compile-exe"]
	if compiler == "" {
		return nil
	}
	exe := make([]string, 0)

	json.Unmarshal([]byte(compiler), &exe)

	if len(exe) < 2 {
		return nil
	}

	pexe, _ := exec.LookPath(exe[0])
	argums := make([]string, 0)
	for _, arg := range exe {
		arg = strings.ReplaceAll(arg, "{source}", scriptm)
		argums = append(argums, arg)
	}
	dir := filepath.Dir(scriptm)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   pexe,
		Args:   argums,
		Dir:    dir,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cmd.Run()

	sout := strings.TrimSpace(stdout.String())
	serr := strings.TrimSpace(stderr.String())
	msg := ""
	if sout != "" {
		msg += sout + "\n"
	}
	if serr != "" {
		msg += serr + "\n"
	}
	if msg == "" {
		return nil
	}

	return fmt.Errorf("compile `%s` with `%s`:\n%s", scriptm, pexe, msg)
}
