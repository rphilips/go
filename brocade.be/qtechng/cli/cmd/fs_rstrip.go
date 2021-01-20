package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"unicode"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var fsRStripCmd = &cobra.Command{
	Use:   "rstrip",
	Short: "rstrips lines files",
	Long: `Each line in the files is right-stripped and a end-of-line is added
Default end-of-line conventions is UNIX-style,
The arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs rstrip cwd=../catalografie`,
	RunE:    fsRStrip,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

// Fwineol windows end-of-line
var Fwineol bool

func init() {
	fsRStripCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsRStripCmd.Flags().BoolVar(&Fwineol, "wineol", false, "Apply MS-Windows end-of-line convention")
	fsRStripCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsRStripCmd)
}

func fsRStrip(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false

	if len(args) == 0 {
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
	if ask && !Fwineol {
		fmt.Print("Windows end-of-line ?         : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fwineol = true
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
		}
	}

	files, err := glob(Fcwd, args, Frecurse, Fpattern)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qerror.ShowResult("", Fjq, err)
			return nil
		}
		msg := make(map[string][]string)
		msg["rstripped"] = files
		Fmsg = qerror.ShowResult(msg, Fjq, nil)
		return nil
	}

	eol := []byte("\n")
	if Fwineol {
		eol = []byte("\r\n")
	}
	fn := func(n int) (interface{}, error) {

		src := files[n]
		// make a copy of the file
		basename := path.Base(src)
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

			eolok := bytes.HasSuffix(buf, eol)
			if !eolok {
				ok = true
			}
			length := len(buf)
			buf = bytes.TrimRightFunc(buf, unicode.IsSpace)
			buf = append(buf, eol...)
			if len(buf) != length {
				ok = true
			}
			_, e := fo.Write(rbuf)
			if e != nil {
				return false, e
			}
		}
		qfs.Rmpath((tmpfile))
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
			changed = append(changed, src)
		}
	}

	msg := make(map[string][]string)
	msg["rstripped"] = changed
	if len(errs) == 0 {
		Fmsg = qerror.ShowResult(msg, Fjq, nil)
	} else {
		Fmsg = qerror.ShowResult(msg, Fjq, qerror.ErrorSlice(errs))
	}
	return nil
}
