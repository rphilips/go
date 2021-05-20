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
	Short: "Executes a sed command",
	Long: `First argument is a sed command. This command is executed on every file
Take care: replacement is done over binary files as well!
The other arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs sed "/remark/d" cwd=../catalografie`,
	RunE:    fsSed,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsSedCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsSedCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsSedCmd)
}

func fsSed(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter sed command       : ")
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
			text = strings.TrimSuffix(text, "\n")
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

	program := strings.NewReader(args[0])
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
	files, err = glob(Fcwd, args[1:], Frecurse, Fpattern, true, false)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
			return nil
		}
		msg := make(map[string][]string)
		msg["sed"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote)
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
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote)
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote)
	}
	return nil
}
