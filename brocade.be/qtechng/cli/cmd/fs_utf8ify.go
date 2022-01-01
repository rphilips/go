package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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

Warning! This command can permanently alter your files.

Some remarks:

	- With no arguments, the AWK program is applied to stdin,
	  output is written to stdout
	- Otherwise, the output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- The arguments are files or directories on which the AWK instruction works
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are taken.
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--replacement' holds the string which replaces the NON-UTF8. If empty, the replacement string stands for U+FFFD`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs utf8ify  . --recurse --cwd=../workspace
qtechng fs utf8ify . --recurse --cwd=../workspace --replacement=PROBLEM
qtechng fs utf8ify --ask`,
	RunE: fsUTF8ify,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsUTF8ifyCmd.Flags().StringVar(&Freplacement, "replacement", "", "Replacement character(s)")
	fsCmd.AddCommand(fsUTF8ifyCmd)
}

func fsUTF8ify(cmd *cobra.Command, args []string) error {
	if Freplacement == "" {
		Freplacement = "\uFFFD"
	}
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"replacement:files:" + Freplacement,
			"ext:awk,files:" + Fext,
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-utf8ify-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Freplacement = argums["replacement"].(string)
		Fext = argums["ext"].(string)
	}

	repl := make([]byte, 0)
	if Freplacement != "" {
		repl = []byte(Freplacement)
	}

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-utf8ify-glob"
			return err
		}
	}
	if len(files) == 0 && len(args) != 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-utf8ify-nofiles")
		return nil
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

	if len(args) == 0 {
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
				Ref:  []string{"fs.utf8ify"},
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
	msg["utf8ify"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-utf8ify-error")
	}
	return nil
}
