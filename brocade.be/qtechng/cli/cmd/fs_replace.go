package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "Replace a string with another one in files",
	Long: `Replace a string with another one in files.
	First argument is the string to search for,
	the second argument is the replacement.

	Warning! This command is very powerful and can permanently alter your files.

	Some remarks:

        - Replacement is done line per line
		- With only two argument, the program is applied to stdin,
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

	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs replace cwd=../catalografie
	qtechng fs replace aap noot /home/tdeneire/tmp --pattern=*.txt`,
	RunE: fsReplace,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsReplaceCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsReplaceCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsReplaceCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsReplaceCmd.Flags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	fsReplaceCmd.Flags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	fsCmd.AddCommand(fsReplaceCmd)
}

func fsReplace(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	if len(args) == 0 {
		fmt.Print("Enter search string: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
	}
	if len(args) == 1 {
		fmt.Print("Enter replacement string: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		args = append(args, text)
	}

	extra, recurse, patterns, utf8only, rexp := qutil.AskArg(args, 2, !Frecurse, len(Fpattern) == 0, !Futf8only, !Fregexp)

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
		if rexp {
			Fregexp = true
		}
	}

	replace := []byte(args[1])
	needle := []byte(args[0])
	sneedle := args[0]
	var rneedle *regexp.Regexp
	var err error

	if Fregexp {
		rneedle, err = regexp.Compile(args[0])
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-invalidregexp")
			return nil
		}
	}

	frepl := func(reader io.Reader, writer io.Writer) (found bool, err error) {
		input := bufio.NewReader(reader)
		for {
			ok := false
			buf, err := input.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				return found, err
			}
			if !Fregexp {
				if bytes.Contains(buf, needle) {
					found = true
					ok = true
				}
			} else {
				ok, _ = regexp.Match(sneedle, buf)
				found = found || ok
			}
			if !ok {
				writer.Write(buf)
			} else {
				if !Fregexp {
					rbuf := bytes.ReplaceAll(buf, needle, replace)
					writer.Write(rbuf)
				} else {
					rbuf := rneedle.ReplaceAll(buf, replace)
					writer.Write(rbuf)
				}
			}
			if err != nil {
				err = nil
				break
			}
		}
		return found, err
	}

	files := make([]string, 0)
	if len(args) > 2 {
		files, err = glob(Fcwd, args[2:], Frecurse, Fpattern, true, false, Futf8only)
		if len(files) == 0 {
			if err != nil {
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-nofiles")
				return nil
			}
		}
	}
	if len(args) == 2 {
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
		frepl(os.Stdin, f)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-nofiles")
			return nil
		}
		msg := make(map[string][]string)
		msg["replaced"] = files
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
		tmpfile, err := qfs.TempFile("", ".repl")
		if err != nil {
			return false, err
		}
		defer qfs.Rmpath(tmpfile)
		f, err := os.Create(tmpfile)

		if err != nil {
			return false, err
		}
		defer f.Close()
		found, err := frepl(in, f)
		f.Close()
		in.Close()
		if err != nil || !found {
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
				Ref:  []string{"fs.replace"},
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
	msg["replaced"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-errors")
	}
	return nil
}
