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
		buf = buf[:0]
		sb.WriteString(`"_$C(`)
		sb.WriteString(prefix)
		sb.WriteString(`)`)
	} else {
		sb.WriteString(`"`)
	}
	x := sb.String()
	if strings.HasPrefix(x, `""_`) {
		x = x[3:]
	}
	return x
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
	rou := qregistry.Registry["m-import-auto-exe"]
	rouparts := make([]string, 0)
	if rou != "" {
		e := json.Unmarshal([]byte(rou), &rouparts)
		if e != nil {
			return nil, fmt.Errorf("Registry value `m-import-auto-exe` is not JSON: `%s`", e.Error())
		}
	} else {
		target := filepath.Join(qregistry.Registry["scratch-dir"], "mumpssinc")
		rouparts = []string{
			"qtechng",
			"fs",
			"store",
			target,
			"--append",
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
