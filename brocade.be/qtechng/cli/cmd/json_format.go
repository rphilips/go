package cmd

import (
	"errors"
	"os"
	"strconv"

	qfs "brocade.be/base/fs"
	qjson "brocade.be/base/json"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var jsonFormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Formats JSON",
	Long: `Formats json files

The arguments are files or directories.
A directory stand for ALL its files.

With 0 arguments, stdin is formatted
These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

Some remarks:

	- With no arguments, the program is applied to stdin,
	  output is written to stdout
	- With more than one argument, the output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- With the '--ask' flag, you can interactively specify the arguments and flags`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng json format . --cwd=../workspace --pattern='*.m'
qtechng json format --ask`,
	RunE: jsonFormat,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Findent = 0

func init() {
	jsonFormatCmd.Flags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	jsonFormatCmd.Flags().IntVar(&Findent, "indent", 2, "Indent length")
	jsonCmd.AddCommand(jsonFormatCmd)
}

func jsonFormat(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"ext:files:" + Fext,
			"indent:files:" + strconv.Itoa(Findent),
		}
		argums, abort := qutil.AskArgs(askfor, Fcwd)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-format-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fext = argums["ext"].(string)
		var err error
		Findent, err = strconv.Atoi(argums["indent"].(string))
		if err != nil {
			Ferrid = "json-format-indent"
			return err
		}
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "json-format-glob"
			return err
		}
	}
	if len(files) == 0 && len(args) != 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-format-nofiles")
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
		err = format(in, src, Fext, Findent)

		if src != "" {
			in.Close()
		}
		return "", err
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var formatted []string
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {

			e := &qerror.QError{
				Ref:  []string{"json-format"},
				File: src,
				Msg:  []string{r.(string) + "->" + errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		if src == "" {
			formatted = append(formatted, "stdin")
		} else {
			formatted = append(formatted, src)
		}
	}

	msg := make(map[string][]string)
	msg["formatted"] = formatted
	if len(errs) == 0 {
		if files[0] != "" {
			Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "json-format-errors")
	}
	return nil
}

func format(in *os.File, src string, ext string, lindent int) error {

	out := os.Stdout
	var err error
	tmpfile := ""
	if src != "" {
		tmpfile, err = qfs.TempFile("", ".format.json")
		if err != nil {
			return err
		}
		defer qfs.Rmpath(tmpfile)
		out, err = os.Create(tmpfile)
		if err != nil {
			return err
		}
		defer out.Close()
	}

	err = qjson.Format(in, out, lindent)
	if err != nil {
		return err
	}

	if src != "" {
		qfs.CopyMeta(src, tmpfile, false)
		err = qfs.CopyFile(tmpfile, src+ext, "=", false)
		if err != nil {
			return err
		}
	}
	return nil
}
