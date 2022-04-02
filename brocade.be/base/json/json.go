package json

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func Format(input io.Reader, output io.Writer, indent int) (err error) {
	if indent < 1 {
		indent = 2
	}

	dec := json.NewDecoder(input)

	level := 0

	written := make(map[int]bool)
	isobject := make(map[int]bool)
	iskey := make(map[int]bool)
	key := make(map[int]string)

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch ty := t.(type) {
		case json.Delim:
			delim := rune(ty)
			sindent := ""
			if delim == '[' || delim == '{' {
				if level > 0 {
					sindent = strings.Repeat(" ", level*indent)
				}
			} else {
				if level > 1 {
					sindent = strings.Repeat(" ", (level-1)*indent)
				}
			}
			if delim == '[' || delim == '{' {
				if written[level] {
					fmt.Fprint(output, ",\n")
				} else {
					if level != 0 {
						fmt.Fprint(output, "\n")
					}
				}
				if isobject[level] {
					fmt.Fprintf(output, "%s%s: ", sindent, key[level])
					iskey[level] = false
					sindent = ""
				}
				fmt.Fprintf(output, "%s%c", sindent, delim)
				level++
				isobject[level] = delim == '{'
				iskey[level] = false
				written[level] = false
				continue
			}
			fmt.Fprintf(output, "\n%s%c", sindent, delim)
			level--
			written[level] = true
		default:

			if isobject[level] && !iskey[level] {
				iskey[level] = true
				b, _ := json.Marshal(ty)
				key[level] = string(b)
				continue
			}
			if written[level] {
				fmt.Fprint(output, ",\n")
			} else {
				fmt.Fprintf(output, "\n")
			}

			sindent := ""
			if level != 0 {
				sindent = strings.Repeat(" ", level*indent)
			}

			if isobject[level] {
				fmt.Fprintf(output, "%s%s: ", sindent, key[level])
				iskey[level] = false
				sindent = ""
			}
			show, _ := json.Marshal(ty)
			fmt.Fprintf(output, "%s%s", sindent, show)
			written[level] = true
		}

	}
	return nil
}
