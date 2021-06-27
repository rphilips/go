package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsNoutf8Cmd = &cobra.Command{
	Use:     "noutf8",
	Short:   "Searches for non-UTF8 byte sequences",
	Long:    `Searches for non-UTF8 byte sequences`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs noutf8 *.m cwd=../catalografie`,
	RunE:    fsNoutf8,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsNoutf8Cmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsNoutf8Cmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsNoutf8Cmd)
}

func fsNoutf8(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false

	if len(args) == 0 {
		ask = true
		for {
			fmt.Print("File/directory          : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 0 {
			return nil
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

	fnoutf8 := func(reader io.Reader) (result [][2]int, err error) {
		var breader *bufio.Reader
		if reader == nil {
			breader = bufio.NewReader(os.Stdin)
		} else {
			breader = bufio.NewReader(reader)
		}
		count := 0
		repl := rune(65533)
		for {
			count++
			line, e := breader.ReadBytes('\n')
			if e != nil && e != io.EOF {
				err = e
				break
			}
			if utf8.Valid(line) && !bytes.ContainsRune(line, repl) {
				if e == io.EOF {
					break
				}
				continue
			}

			good := strings.ToValidUTF8(string(line), string(repl))
			parts := strings.SplitN(good, string(repl), -1)

			total := ""
			for c, part := range parts {
				if c == len(parts) {
					break
				}
				total += part + "\n"
				result = append(result, [2]int{count, len([]rune(total))})
			}
		}
		return
	}

	if len(args) == 2 && args[1] == "-" {
		result, err := fnoutf8(nil)
		if len(result) == 0 {
			result = nil
		}
		Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	files := make([]string, 0)
	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["noutf8"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	fn := func(n int) (interface{}, error) {
		src := files[n]

		in, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		defer in.Close()
		result, err := fnoutf8(in)
		return result, err
	}
	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	m := make(map[string][][2]int)
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.noutf8"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		m[src] = r.([][2]int)
	}
	if len(errs) == 0 {
		errs = nil
	}

	msg := make(map[string]map[string][][2]int)
	msg["noutf8"] = m
	Fmsg = qreport.Report(msg, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
