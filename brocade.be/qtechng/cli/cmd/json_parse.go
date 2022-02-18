package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var jsonParseCmd = &cobra.Command{
	Use:   "parse",
	Short: "Parses JSON",
	Long: `Parses json files

The arguments are files or directories.
A directory stand for ALL its files.

With 0 arguments, stdin is parsed
These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

	Some remarks:

	- Search is done line per line
	- With the '--ask' flag, you can interactively specify the arguments and flags`,

	Args: cobra.MinimumNArgs(0),
	Example: `qtechng json parse . --cwd=../workspace --pattern='*.m'
	qtechng json parse --ask`,
	RunE: jsonParse,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	jsonParseCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	jsonParseCmd.Flags().StringVar(&Fsearch, "search", "", "Search for")
	jsonParseCmd.Flags().BoolVar(&Futf8only, "utf8only", false, "Only UTF-8 encodes files are considered")
	jsonCmd.AddCommand(jsonParseCmd)
}

func jsonParse(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor, Fcwd)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-parse-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "json-parse-glob"
			return err
		}
	}
	if len(args) != 0 && len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-parse-nofiles")
		return nil
	}

	if len(files) == 0 {
		files = append(files, "")
	}

	fn := func(n int) (interface{}, error) {
		src := files[n]
		var in *os.File
		var err error
		if src != "" {
			in, err = os.Open(src)
		} else {
			in = os.Stdin
		}
		if err != nil {
			return nil, err
		}

		if src != "" {
			defer in.Close()
		}

		dec := json.NewDecoder(in)
		after := make([]json.Token, 0)
		start := 0
		place := -1
		for {
			t, err := dec.Token()
			if t != nil {
				if start < 36 {
					after = append(after, t)
					place++
				} else {
					place = start % 36
					after[place] = t
				}
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				msg := ""
				if place < 1+len(after) {
					for _, x := range after[place+1:] {
						msg += fmt.Sprintf("%v", x)
					}
				}
				for i, x := range after {
					if i > place {
						break
					}
					msg += fmt.Sprintf("%v", x)
				}
				return msg, err
			}
			start++
		}
		return "", nil
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var parsed []string
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {

			e := &qerror.QError{
				Ref:  []string{"json-parse"},
				File: src,
				Msg:  []string{r.(string) + "->" + errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		if src == "" {
			parsed = append(parsed, "stdin")
		} else {
			parsed = append(parsed, src)
		}
	}

	msg := make(map[string][]string)
	msg["parsed"] = parsed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-parse-errors")
	}
	return nil
}
