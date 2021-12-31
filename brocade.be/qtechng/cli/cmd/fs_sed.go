package cmd

import (
	"errors"
	"io"
	"log"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	qsed "github.com/rwtodd/Go.Sed/sed"
	"github.com/spf13/cobra"
)

var fsSedCmd = &cobra.Command{
	Use:   "sed",
	Short: "Execute a sed command",
	Long: `Executes a *sed* command on files.
The first argument is a sed command.
See: https://en.wikipedia.org/wiki/Sed

Warning! This command is very powerful and can permanently alter your files.

Some remarks:

	- With no arguments, the sed program is applied to stdin,
		output is written to stdout
	- Otherwise, the output is written to the same file with the '--ext'
		flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- The arguments are files or directories on which the sed instruction works
		(use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- The '--sed' flag contains the sed statement
	- The '--isfile' flag specifies that the '--sed' flag is the name of a file
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs sed '/remark/d' . cwd=../workspace --patern='*.txt'
qtechng fs sed --ask`,
	RunE: fsSed,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}
var Fsed = ""

func init() {
	fsSedCmd.Flags().BoolVar(&Fisfile, "isfile", false, "Is this an AWK file?")
	fsSedCmd.Flags().StringVar(&Fsed, "sed", "", "sed command")
	fsCmd.AddCommand(fsSedCmd)
}

func fsSed(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"sed::" + Fsed,
			"isfile:sed:" + qutil.UnYes(Fisfile),
			"files:sed",
			"recurse:sed,files:" + qutil.UnYes(Frecurse),
			"patterns:sed,files:",
			"utf8only:sed,files:" + qutil.UnYes(Futf8only),
			"ext:sed,files:" + Fext,
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-abort")
			return nil
		}
		Fsed = argums["sed"].(string)
		args = argums["files"].([]string)
		Fisfile = argums["isfile"].(bool)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fext = argums["ext"].(string)
	}
	if Fsed == "" {
		Fmsg = qreport.Report(nil, errors.New("missing sed statement"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-cmd")
		return nil
	}

	var program io.Reader
	if Fisfile {
		var err error
		fl, err := os.Open(Fsed)
		if err != nil {
			Ferrid = "fs-sed-sedfile"
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
		defer fl.Close()
		program = fl
	} else {
		program = strings.NewReader(Fsed)
	}
	engine, err := qsed.New(program)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-invalidsed")
		return nil
	}

	fsed := func(reader io.Reader, writer io.Writer) error {
		if reader == nil {
			reader = os.Stdin
		}
		if writer == nil {
			writer = os.Stdout
		}
		_, err := io.Copy(writer, engine.Wrap(reader))
		return err
	}

	files := make([]string, 0)
	if len(args) != 0 {
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-sed-glob"
			return err
		}
		if len(files) == 0 {
			Fmsg = qreport.Report(nil, errors.New("no files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-nofiles")
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
		return fsed(os.Stdin, f)
	}

	fn := func(n int) (interface{}, error) {
		src := files[n]

		in, err := os.Open(src)
		if err != nil {
			return false, err
		}
		defer in.Close()
		tmpfile, err := qfs.TempFile("", ".sed")
		if err != nil {
			return false, err
		}
		defer qfs.Rmpath(tmpfile)

		f, err := os.Create(tmpfile)

		if err != nil {
			return false, err
		}
		defer f.Close()
		err = fsed(in, f)
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
				Ref:  []string{"fs.sed"},
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
	msg["sed"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-errors")
	}
	return nil
}
