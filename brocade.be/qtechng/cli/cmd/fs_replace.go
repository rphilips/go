package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsReplaceCmd = &cobra.Command{
	Use:   "replace",
	Short: "replaces a string to another string in files",
	Long: `First argument is the string to search for, the second argument is the replacement.
Replacement is done line per line.
Take care: replacement is done over binary files as well!
The other arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs replace cwd=../catalografie`,
	RunE:    fsReplace,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsReplaceCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsReplaceCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsReplaceCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsReplaceCmd)
}

func fsReplace(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter search string     : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
	}
	if len(args) == 1 {
		ask = true
		fmt.Print("Enter replacement string: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		args = append(args, text)
	}

	if len(args) == 2 {
		ask = true
		for {
			fmt.Print("File/directory          : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 2 {
			return nil
		}
	}
	if ask && !Fregexp {
		fmt.Print("Regexp ?                : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fregexp = true
		}
	}

	if ask && !Frecurse {
		fmt.Print("Recurse ?               : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
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
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			Fpattern = append(Fpattern, text)
			if text == "*" {
				break
			}
		}
	}

	replace := []byte(args[1])
	needle := []byte(args[0])
	sneedle := args[0]
	var rneedle *regexp.Regexp
	var err error
	files := make([]string, 0)
	if Fregexp {
		rneedle, err = regexp.Compile(args[0])
	}
	if err == nil {
		files, err = glob(Fcwd, args[2:], Frecurse, Fpattern, true, false, true)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["replaced"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	fn := func(n int) (interface{}, error) {

		src := files[n]
		in, err := os.Open(src)
		if err != nil {
			return false, err
		}
		input := bufio.NewReader(in)
		defer in.Close()

		ok := false
		for {
			// read a chunk
			buf, err := input.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				return false, err
			}
			if !Fregexp {
				if bytes.Contains(buf, needle) {
					ok = true
				}
			} else {
				ok, _ = regexp.Match(sneedle, buf)
			}
			if ok {
				break
			}
			if err != nil {
				break
			}
		}
		if !ok {
			return false, nil
		}
		in.Close()
		// make a copy of the file
		basename := filepath.Base(src)
		tmpfile, err := qfs.TempFile("", "fs-replace."+basename+".")
		if err != nil {
			return false, err
		}
		err = qfs.CopyFile(src, tmpfile, "", false)
		if err != nil {
			return false, err
		}
		in, err = os.Open(tmpfile)
		if err != nil {
			return false, err
		}
		input = bufio.NewReader(in)
		defer in.Close()

		// open output file
		fo, err := os.Create(src)
		if err != nil {
			return false, err
		}
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		var rbuf []byte
		for {
			// read a chunk
			buf, err := input.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				return false, err
			}
			if !Fregexp {
				rbuf = bytes.ReplaceAll(buf, needle, replace)
			} else {
				rbuf = rneedle.ReplaceAll(buf, replace)
			}
			_, e := fo.Write(rbuf)
			if e != nil {
				return false, e
			}
			if err != nil {
				break
			}
		}
		qfs.Rmpath(tmpfile)
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
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
