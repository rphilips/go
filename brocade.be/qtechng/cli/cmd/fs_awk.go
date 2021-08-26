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
	- If the second argument is '-', the AWK program is applied to stdin.
	- If no arguments are given, the command asks for arguments.
	- The other arguments: at least one file or directory are to be specified.
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames`,

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
	fsAWKCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsAWKCmd)
}

func fsAWK(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter AWK command       : ")
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

	program := args[0]
	if Fisfile {
		var err error
		body, err := qfs.Fetch(program)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		program = string(body)
	}
	_, err := qawkp.ParseProgram([]byte(program), nil)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	fawk := func(reader io.Reader, writer io.Writer) error {
		return qawk.Exec(program, " ", reader, writer)
	}

	if len(args) == 2 && args[1] == "-" {

		if Fstdout != "" {
			f, err := os.Create(Fstdout)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			return fawk(nil, f)
		}
		return fawk(nil, nil)
	}

	files := make([]string, 0)
	if len(args) != 2 || args[1] != "-" {
		files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false, true)
	}
	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["awk"] = files
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
				Ref:  []string{"fs.awk"},
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
	msg["awk"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
