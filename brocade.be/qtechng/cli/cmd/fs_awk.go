package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	qawk "github.com/benhoyt/goawk/interp"
	qawkp "github.com/benhoyt/goawk/parser"
	"github.com/spf13/cobra"
)

var fsAWKCmd = &cobra.Command{
	Use:   "awk",
	Short: "Execute an AWK command",
	Long: `Executes an AWK command on files.
The first argument is an AWK command.
See: https://en.wikipedia.org/wiki/AWK

The files are given as input to the AWK statements.
The result - what appears on stdout - replaces the original file!

Warning! This command is very powerful and can permanently alter your files.

Some remarks:

    - This command is executed only on files which are deemed valid UTF-8 files.
	- With only one argument, the AWK program is applied to stdin,
	  output is written to stdout
	- With more than one argument, the output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- If no arguments are given, the command asks for arguments.
	- The other arguments: at least one file or directory are to be specified.
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content`,

	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs awk '{print $1}' . --cwd=../catalografie`,
	RunE:    fsAWK,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsAWKCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsAWKCmd.Flags().BoolVar(&Fisfile, "isfile", false, "Is this an AWK file?")
	fsAWKCmd.Flags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	fsAWKCmd.Flags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	fsAWKCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsAWKCmd)
}

func fsAWK(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	if len(args) == 0 {
		fmt.Print("Enter AWK command: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
	}
	extra, recurse, patterns, utf8only, _ := qutil.AskArg(args, 1, !Frecurse, len(Fpattern) == 0, !Futf8only, false)

	if len(extra) != 0 {
		args = append(args, extra...)
		if recurse {
			Frecurse = true
		}
		if len(patterns) != 0 {
			Fpattern = patterns
		}
		if utf8only {
			Futf8only = true
		}
	}

	program := args[0]

	if Fisfile {
		var err error
		body, err := qfs.Fetch(program)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-invalidfile")
			return nil
		}
		program = string(body)
	}
	_, err := qawkp.ParseProgram([]byte(program), nil)

	if err != nil {

		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-invalidawk")
		return nil
	}

	fawk := func(reader io.Reader, writer io.Writer) error {
		return qawk.Exec(program, " ", reader, writer)
	}

	files := make([]string, 0)
	if len(args) > 1 {
		files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-awk-glob"
			return err
		}
	}

	if len(args) == 1 {
		var f *os.File = nil
		if Fstdout != "" {
			f, err = os.Create(Fstdout)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			f = os.Stdout
		}
		if f != nil {
			defer f.Close()
		}
		return fawk(os.Stdin, f)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-nofiles")
			return nil
		}
		msg := make(map[string][]string)
		msg["awk"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	fn := func(n int) (interface{}, error) {
		src := files[n]

		in, err := os.Open(src)
		if err != nil {
			return false, err
		}
		defer in.Close()
		tmpfile, err := qfs.TempFile("", ".awk")
		if err != nil {
			return false, err
		}
		defer qfs.Rmpath(tmpfile)

		f, err := os.Create(tmpfile)

		if err != nil {
			return false, err
		}
		defer f.Close()
		err = fawk(in, f)
		f.Close()
		in.Close()
		if err != nil {
			return false, err
		}
		qfs.CopyMeta(src, tmpfile, false)
		err = qfs.CopyFile(tmpfile, src+Fext, "=", false)
		if err != nil {
			return false, err
		}
		return true, nil
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var changed []string
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.awk"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		if r.(bool) {
			changed = append(changed, src+Fext)
		}
	}

	msg := make(map[string][]string)
	msg["awk"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-errors")
	}
	return nil
}
