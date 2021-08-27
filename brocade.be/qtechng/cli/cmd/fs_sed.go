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

		- This command is executed only on files which are deemed valid UTF-8 files.
		- If the second argument is '-', the sed program is applied to stdin.
		- If no arguments are given, the command asks for arguments.
		- The other arguments: at least one file or directory are to be specified.
		  (use '.' to indicate the current working directory)
		- If an argument is a directory, all files in that directory are handled.
		- The '--recurse' flag recursively traverses the subdirectories of the argument directories.
		- The '--pattern' flag builds a list of acceptable patterns on the basenames.`,
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
	fsCmd.AddCommand(fsSedCmd)
}

func fsSed(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter sed command       : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}
		args = append(args, text)
	}
	if len(args) == 1 {
		ask = true
		for {
			fmt.Print("File/directory          : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSpace(text)
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 1 {
			return nil
		}
	}

	if ask && !Frecurse {
		fmt.Print("Recurse ?               : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
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
			text = strings.TrimSpace(text)
			if text == "" {
				break
			}
			Fpattern = append(Fpattern, text)
			if text == "*" {
				break
			}
		}
	}
	var program io.Reader
	if Fisfile {
		var err error
		fl, err := os.Open(args[0])
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		defer fl.Close()
		program = fl
	} else {
		program = strings.NewReader(args[0])
	}
	engine, err := qsed.New(program)

	if err != nil {
		return err
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

	if len(args) == 2 && args[1] == "-" {

		if Fstdout != "" {
			f, err := os.Create(Fstdout)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			return fsed(nil, f)
		}
		return fsed(nil, nil)
	}

	files := make([]string, 0)
	files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false, true)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["sed"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	if err != nil {
		return err
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
		err = qfs.CopyFile(tmpfile, src, "=", false)
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
			changed = append(changed, src)
		}
	}

	msg := make(map[string][]string)
	msg["sed"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
