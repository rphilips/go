package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsGrepCmd = &cobra.Command{
	Use:   "grep",
	Short: "Execute *grep* on files",
	Long: `Searches files for content

	The arguments are files or directories.
	A directory stand for ALL files.

	These argument scan be expanded/restricted by using the flags:

		- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
		- The '--pattern' flag builds a list of acceptable patterns on the basenames
		- The '--utf8only' flag restricts to files with UTF-8 content

	The search action on the lines in the argument are guided by:

		- The '--search' flag gives the string to search for in each line of the argument
		- The '--regexp' flag indicates if the '--search' flag is a regular expression


	Some remarks:

		- Search is done line per line
		- With the '--ask' flag, you can interactively specify the arguments and flags`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs grep UDuser ../catalografie`,
	RunE:    fsGrep,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsGrepCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsGrepCmd.Flags().StringVar(&Fsearch, "search", "", "Search for")
	fsGrepCmd.Flags().BoolVar(&Ftolower, "tolower", false, "Lowercase before grepping")
	fsCmd.AddCommand(fsGrepCmd)
}

func fsGrep(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"search::" + Fsearch,
			"regexp:search:" + qutil.UnYes(Fregexp),
			"tolower:search:" + qutil.UnYes(Ftolower),
			"files:search",
			"recurse:search,files:" + qutil.UnYes(Frecurse),
			"patterns:search,files:",
			"utf8only:search,files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-abort")
			return nil
		}
		Fsearch = argums["search"].(string)
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fregexp = argums["regexp"].(bool)
		Ftolower = argums["tolower"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}
	if Fsearch == "" {
		Fmsg = qreport.Report(nil, errors.New("search string is empty"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-grep-search")
		return nil
	}
	if Fregexp {
		var err error
		_, err = regexp.Compile(Fsearch)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-grep-invalidregexp")
			return nil
		}
	}
	needle := []byte(Fsearch)
	sneedle := Fsearch

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-grep-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-grep-nofiles")
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
				t := qutil.FileURL(src, "", lin.lineno) + " " + strings.TrimSpace(string(lin.content))
				grep = append(grep, t)
			}
		}
	}

	msg := make(map[string][]string)
	msg["grepped"] = grep
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
