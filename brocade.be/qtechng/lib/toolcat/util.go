package toolcat

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	qclip "brocade.be/clipboard"
	qutil "brocade.be/qtechng/lib/util"
	qyaml "gopkg.in/yaml.v3"
)

func Display(outfile string, cwd string, obj fmt.Stringer, signature string, indent string, after string, replacements map[string]string, clip bool, withdelim bool) (out string, err error) {
	if clip {
		qclip.WriteAll("")
	}
	output := os.Stdout
	if outfile != "" {
		f, err := os.Create(qutil.AbsPath(outfile, cwd))
		if err != nil {
			return "", err
		}
		output = f
		defer output.Close()
	}
	sout := strings.TrimSpace(obj.String())
	lines := strings.SplitN(sout, "\n", -1)
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			lines[i] = ""
			continue
		}
		lines[i] = indent + strings.TrimRightFunc(line, unicode.IsSpace)
	}
	sout = strings.Join(lines, "\n")
	pdelim := `r'''` + "\n"
	adelim := `'''` + "\n"
	if strings.Contains(sout, adelim) {
		pdelim = `r"""` + "\n"
		adelim = `"""` + "\n"
	}
	if signature != "" {
		signature = strings.TrimSpace(signature) + "\n"
	}
	if !withdelim {
		pdelim = ""
		adelim = ""
		indent = ""
	}
	sout = signature + indent + pdelim + sout + "\n" + indent + adelim + "\n" + after
	for key, value := range replacements {
		sout = strings.ReplaceAll(sout, key, value)
	}
	fmt.Fprintln(output, sout)
	if clip {
		qclip.WriteAll(sout)
	}
	return sout, nil
}

func yaml(i interface{}) string {
	b, _ := qutil.Yaml(i)
	return strings.ReplaceAll(string(b), "\n\n\n", "\n\n")
}

func worker(content []string, name string, fn func(string) string, node *qyaml.Node) {
	if len(content) == 0 {
		return
	}
	ok := make(map[string]bool)
	qts := make([]string, len(content))
	i := 0
	for _, qt := range content {
		if fn != nil {
			qt = fn(qt)
		}
		qt = strings.TrimSpace(qt)
		if ok[qt] {
			continue
		}
		ok[qt] = true
		qts[i] = qt
		i++
	}
	sort.Strings(qts)
	node.Content = append(node.Content,
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: name,
			Tag:   "!!str",
		},
		&qyaml.Node{
			Kind:  qyaml.ScalarNode,
			Value: strings.Join(qts, ", "),
			Tag:   "!!str",
		})
}
