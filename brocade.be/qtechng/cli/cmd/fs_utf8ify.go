package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode/utf8"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var Freplacement string
var fsUTF8ifyCmd = &cobra.Command{
	Use:   "utf8ify",
	Short: "Transforms to UTF-8 inplace",
	Long: `Replace NON UTF-8 sequences with a replacement string.

Warning! This command is very powerful and can permanently alter your files.

Some remarks:

	- If the second argument is '-', the UTF8ify program is applied to stdin.
	- If no arguments are given, the command asks for arguments.
	- The other arguments: at least one file or directory are to be specified.
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames`,

	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs utf8ify  . --cwd=../catalografie`,
	RunE:    fsUTF8ify,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsUTF8ifyCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsUTF8ifyCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsUTF8ifyCmd.Flags().StringVar(&Freplacement, "replacement", "", "Replacement character(s)")
	fsCmd.AddCommand(fsUTF8ifyCmd)
}

func fsUTF8ify(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false

	if len(args) == 0 {
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
	repl := make([]byte, 0)
	if Freplacement != "" {
		repl = []byte(qutil.Joiner(Freplacement))
	}

	utf8ify := func(reader io.Reader, repl []byte) (changed bool, valid []byte, err error) {
		var breader *bufio.Reader
		if reader == nil {
			breader = bufio.NewReader(os.Stdin)
		} else {
			breader = bufio.NewReader(reader)
		}
		data, err := ioutil.ReadAll(breader)
		if err != nil {
			return
		}
		if utf8.Valid(data) {

			return changed, data, nil
		}
		valid = bytes.ToValidUTF8(data, repl)
		changed = true
		return
	}

	if len(args) == 1 && args[0] == "-" {
		_, valid, _ := utf8ify(nil, repl)
		if Fstdout != "" {
			f, err := os.Create(Fstdout)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			_, err = f.Write(valid)
			return err
		} else {
			fmt.Print(string(valid))
		}
		return nil
	}

	files := make([]string, 0)
	var err error
	if len(args) != 0 && args[0] != "-" {
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, false)
	}
	if err != nil {
		return err
	}
	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["utf8ify"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	fn := func(n int) (interface{}, error) {
		src := files[n]

		in, err := os.Open(src)
		if err != nil {
			return false, err
		}
		defer in.Close()

		if err != nil {
			return false, err
		}
		changed, valid, err := utf8ify(in, repl)
		in.Close()
		if err != nil {
			return false, err
		}
		if !changed {
			return false, nil
		}
		tmpfile, err := qfs.TempFile("", ".utf8ify")
		if err != nil {
			return false, err
		}
		defer qfs.Rmpath(tmpfile)
		qfs.Store(tmpfile, valid, "temp")
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
				Ref:  []string{"fs.utf8ify"},
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
	msg["utf8ify"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
