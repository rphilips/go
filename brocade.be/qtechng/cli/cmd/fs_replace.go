package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"regexp"

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

Warning! This command is very powerful and can permanently alter your files.

The arguments are files or directories.
A directory stand for ALL files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

The search/replace action on the lines in the argument are guided by:

    - The '--search' flag gives the string to search for in the abspath of the argument
    - The '--replace' flag replaces the 'search' part
	- The '--regexp' flag indicates if the '--search' flag is a regular expression


Some remarks:

	- Replacement is done line per line
	- With no arguments, the program is applied to stdin,
	  output is written to stdout
	- With more than one argument, the output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- With the '--ask' flag, you can interactively specify the arguments and flags`,

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
	fsReplaceCmd.Flags().StringVar(&Fsearch, "search", "", "Search for")
	fsReplaceCmd.Flags().StringVar(&Freplace, "replace", "", "Replace with")
	fsCmd.AddCommand(fsReplaceCmd)
}

func fsReplace(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"search::" + Fsearch,
			"regexp:search:" + qutil.UnYes(Fregexp),
			"replace:search:" + Freplace,
			"files:search",
			"recurse:search,files:" + qutil.UnYes(Frecurse),
			"patterns:search,files:",
			"utf8only:search,files:" + qutil.UnYes(Futf8only),
			"ext:search,files:" + Fext,
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-abort")
			return nil
		}
		Fsearch = argums["search"].(string)
		Freplace = argums["replace"].(string)
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fregexp = argums["regexp"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fext = argums["ext"].(string)
	}
	if Fsearch == "" {
		Fmsg = qreport.Report(nil, errors.New("search string is empty"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-search")
		return nil
	}
	var rneedle *regexp.Regexp
	if Fregexp {
		var err error
		rneedle, err = regexp.Compile(Fsearch)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-invalidregexp")
			return nil
		}
	}
	needle := []byte(Fsearch)
	sneedle := Fsearch
	replace := []byte(Freplace)

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
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-replace-glob"
			return err
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-replace-nofiles")
		return nil
	}
	if len(args) == 0 {
		var f *os.File = nil
		if Fstdout != "" {
			var err error
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
			changed = append(changed, src+Fext)
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
