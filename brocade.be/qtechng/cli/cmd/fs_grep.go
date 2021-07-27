package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsGrepCmd = &cobra.Command{
	Use:   "grep",
	Short: "Searches for a string in files",
	Long: `First argument is the string to search for, 
Search is done line per line.
Take care: search is done over binary files as well!
The other arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs grep cwd=../catalografie`,
	RunE:    fsGrep,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsGrepCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsGrepCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsGrepCmd.Flags().BoolVar(&Ftolower, "tolower", false, "Lowercase before grepping")
	fsGrepCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsGrepCmd)
}

func fsGrep(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter search string     : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
	}

	if len(args) == 1 {
		ask = true
		for {
			fmt.Print("File/directory        : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 1 {
			return nil
		}
	}
	if ask && !Fregexp {
		fmt.Print("Regexp ?                : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fregexp = true
		}
	}

	if ask && !Frecurse {
		fmt.Print("Recurse ?               : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Frecurse = true
		}
	}

	if ask && len(Fpattern) == 0 {
		for {
			fmt.Print("Pattern on basename     : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			Fpattern = append(Fpattern, text)
		}
	}

	needle := []byte(args[0])
	sneedle := args[0]
	var err error
	files := make([]string, 0)
	if Fregexp {
		_, err = regexp.Compile(args[0])
	}
	if err == nil {
		files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false, true)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["grepped"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	type line struct {
		lineno  int
		content []byte
	}

	fn := func(n int) (interface{}, error) {
		lines := make([]line, 0)
		src := files[n]
		in, err := os.Open(src)
		if err != nil {
			return lines, err
		}
		input := bufio.NewReader(in)
		defer in.Close()

		lineno := 0
		for {
			ok := false
			// read a chunk
			buf, err := input.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				return lines, err
			}
			if Ftolower {
				buf = bytes.ToLower(buf)
			}
			lineno++
			if !Fregexp {
				if bytes.Contains(buf, needle) {
					ok = true
				}
			} else {
				ok, _ = regexp.Match(sneedle, buf)
			}
			if ok {
				lines = append(lines, line{lineno, buf})
			}
			if err != nil {
				break
			}
		}
		return lines, nil
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var grep []string
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.grep"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		lins := r.([]line)
		if len(lins) != 0 {
			for _, lin := range lins {
				t := src + ":" + strconv.Itoa(lin.lineno) + ":" + string(lin.content)
				grep = append(grep, t)
			}
		}
	}

	msg := make(map[string][]string)
	msg["grepped"] = grep
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
