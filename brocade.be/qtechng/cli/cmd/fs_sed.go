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

		- With only one argument, the sed program is applied to stdin,
	      output is written to stdout
		- With more than one argument, the output is written to the same file with the '--ext'
	      flag added to the name. (If '--ext' is empty, the file is modified inplace.)
		- If no arguments are given, the command asks for arguments.
		- The other arguments: at least one file or directory are to be specified.
		  (use '.' to indicate the current working directory)
		- If an argument is a directory, all files in that directory are handled.
		- The '--recurse' flag recursively traverses the subdirectories of the argument directories.
		- The '--pattern' flag builds a list of acceptable patterns on the basenames
		- The '--utf8only' flag restricts to files with UTF-8 content.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs sed '/remark/d' cwd=../catalografie`,
	RunE:    fsSed,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsSedCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsSedCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsSedCmd.Flags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	fsSedCmd.Flags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	fsCmd.AddCommand(fsSedCmd)
}

func fsSed(cmd *cobra.Command, args []string) error {
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
	var program io.Reader
	if Fisfile {
		var err error
		fl, err := os.Open(args[0])
		if err != nil {
			Ferrid = "fs-sed-sedfile"
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
		defer fl.Close()
		program = fl
	} else {
		program = strings.NewReader(args[0])
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
	if len(args) > 1 {
		files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-sed-glob"
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
		return fsed(os.Stdin, f)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-sed-nofiles")
			return nil
		}
		msg := make(map[string][]string)
		msg["sed"] = files
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
