package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes files",
	Long: `The arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.
Use the delete flag if the original files should be deleted
`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs delete *.bak`,
	RunE:    fsDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsDeleteCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsDeleteCmd.Flags().BoolVar(&Fconfirm, "confirm", false, "Ask the first time for confirmation")
	fsDeleteCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsDeleteCmd)
}

func fsDelete(cmd *cobra.Command, args []string) error {
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
		}
	}

	if ask && !Fconfirm {
		fmt.Print("Confirm first time ?    : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fconfirm = true
		}
	}

	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false)

	if Fconfirm && len(files) != 0 {
		Fconfirm = false
		fmt.Printf("Delete `%s`: <n>", files[0])
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		ok := false
		if strings.ContainsAny(text, "jJyY1tT") {
			ok = true
		}
		if !ok {
			files = make([]string, 0)
		}
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fsilent)
			return nil
		}
		msg := make(map[string][]string)
		msg["deleted"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}

	fn := func(n int) (interface{}, error) {

		src := files[n]
		err := qfs.Rmpath(src)
		return src, err
	}

	errs := make([]error, 0)
	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var changed []string
	for i, src := range resultlist {
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.delete"},
				File: src.(string),
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		changed = append(changed, src.(string))
	}

	msg := make(map[string][]string)
	msg["deleted"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fsilent)
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fsilent)
	}
	return nil
}
