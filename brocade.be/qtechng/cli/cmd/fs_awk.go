package cmd

import (
	"errors"
	"io"
	"log"
	"os"

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

Warning! This command is very powerful and can permanently alter your files.

Some remarks:

	- With no arguments, the AWK program is applied to stdin,
	  output is written to stdout
	- Otherwise, the output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- The arguments are files or directories on which the AWK instruction works
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- The '--awk' flag contains the AWK statement
	- The '--isfile' flag specifies that the '--awk' flag is the name of a file
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content`,

	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs awk . --awk='{print $1}' --cwd=../workspace --recurse --pattern='*.txt'
qtechng fs awk  --awk='{print $1}' --stdout=result.txt
qtechng fs awk f1.txt f2.txt --awk=myprog.awk --cwd=../workspace --isfile
qtechng fs awk --ask`,
	RunE: fsAWK,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Fawk = ""

func init() {
	fsAWKCmd.Flags().BoolVar(&Fisfile, "isfile", false, "Is this an AWK file?")
	fsAWKCmd.Flags().StringVar(&Fawk, "awk", "", "AWK command")
	fsCmd.AddCommand(fsAWKCmd)
}

func fsAWK(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"awk::" + Fawk,
			"isfile:awk:" + qutil.UnYes(Fisfile),
			"files:awk",
			"recurse:awk,files:" + qutil.UnYes(Frecurse),
			"patterns:awk,files:",
			"utf8only:awk,files:" + qutil.UnYes(Futf8only),
			"ext:awk,files:" + Fext,
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-abort")
			return nil
		}
		Fawk = argums["awk"].(string)
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fisfile = argums["isfile"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fext = argums["ext"].(string)
	}
	if Fawk == "" {
		Fmsg = qreport.Report(nil, errors.New("missing AWK statement"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-cmd")
		return nil
	}

	program := Fawk

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
	if len(args) != 0 {
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-awk-glob"
			return err
		}
		if len(files) == 0 {
			Fmsg = qreport.Report(nil, errors.New("no files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-awk-nofiles")
			return nil
		}
	}

	if len(args) == 0 {
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
