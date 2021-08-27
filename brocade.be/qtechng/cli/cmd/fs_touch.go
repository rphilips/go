package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsTouchCmd = &cobra.Command{
	Use:   "touch",
	Short: "Touch files",
	Long: `The last modification time of the file(s) is changed to the current moment.
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs touch cwd=../catalografie`,
	RunE:    fsTouch,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsTouchCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsTouchCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsTouchCmd)
}

func fsTouch(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		for {
			fmt.Print("File/directory        : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 0 {
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
			if text == "*" {
				break
			}
		}
	}
	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false, false)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["touched"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	h := time.Now().Local()
	fn := func(n int) (interface{}, error) {
		src := files[n]
		et := os.Chtimes(src, h, h)
		return src, et
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var touched []string
	for i := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.touch"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		touched = append(touched, src)
	}

	msg := make(map[string][]string)
	msg["touched"] = touched
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
