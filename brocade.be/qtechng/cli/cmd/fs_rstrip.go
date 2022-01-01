package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"unicode"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsRStripCmd = &cobra.Command{
	Use:   "rstrip",
	Short: "Execute rstrip on file lines",
	Long: `Execute rstrip on file lines


Each line in the files is right-stripped and an appropriate end-of-line is added.

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

Some remarks:

    - The output is written to the same file with the '--ext'
	  flag added to the name. (If '--ext' is empty, the file is modified inplace.)
	- The '--unix' flag ensures Unix EOL convention
	- The '--windows' flag ensures Windows EOL convention
	- Without these flags, the EOL convention is based line-per-line
	- With the '--ask' flag, you can interactively specify the arguments and flags`,

	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs rstrip myfile.m --unix --cwd=../workspace
qtechng fs rstrip --ask`,
	RunE: fsRStrip,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

// Fwineol windows end-of-line
var Fwindows bool
var Funix bool

func init() {
	fsRStripCmd.Flags().BoolVar(&Fwindows, "windows", false, "Apply MS-Windows end-of-line convention")
	fsRStripCmd.Flags().BoolVar(&Funix, "unix", false, "Apply Unix end-of-line convention")
	fsCmd.AddCommand(fsRStripCmd)
}

func fsRStrip(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"windows:files:" + qutil.UnYes(Fwindows),
			"unix:files,!windows:" + qutil.UnYes(Fwindows),
			"ext:awk,files:" + Fext,
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-rstrip-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fwindows = argums["windows"].(bool)
		Funix = argums["unix"].(bool)
		Fext = argums["ext"].(string)
	}

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-rstrip-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-rstrip-nofiles")
		return nil
	}
	whateol := true
	eol := []byte{}
	if Fwindows {
		eol = []byte("\r\n")
		whateol = false
	}
	if Funix {
		eol = []byte("\n")
		whateol = false
	}
	fn := func(n int) (interface{}, error) {

		src := files[n]
		// make a copy of the file
		basename := filepath.Base(src)
		tmpfile, err := qfs.TempFile("", "fs-rstrip."+basename+".")
		if err != nil {
			return false, err
		}
		err = qfs.CopyFile(src, tmpfile, "", false)
		if err != nil {
			return false, err
		}
		in, err := os.Open(tmpfile)
		if err != nil {
			return false, err
		}
		input := bufio.NewReader(in)
		defer in.Close()

		// open output file
		fo, err := os.Create(src + Fext)
		if err != nil {
			return false, err
		}
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
		ok := false
		for {
			// read a chunk
			buf, err := input.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				return false, err
			}
			if err == io.EOF {
				if len(buf) == 0 {
					break
				}
				length := len(buf)

				buf = bytes.TrimRightFunc(buf, unicode.IsSpace)
				if len(buf) != length {
					ok = true
				}
				_, e := fo.Write(buf)
				if e != nil {
					return false, e
				}
				return ok, nil
			}
			if whateol {
				eol = []byte("\n")
				if bytes.HasSuffix(buf, []byte("\r\n")) {
					eol = []byte("\r\n")
				}
			}
			length := len(buf)
			buf = bytes.TrimRightFunc(buf, unicode.IsSpace)
			buf = append(buf, eol...)
			if len(buf) != length {
				ok = true
			}
			_, e := fo.Write(buf)
			if e != nil {
				return false, e
			}
		}
		qfs.Rmpath(tmpfile)
		return ok, nil
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var changed []string
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.rstrip"},
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
	msg["rstripped"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
