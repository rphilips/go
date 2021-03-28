package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var fsSetpropertyCmd = &cobra.Command{
	Use:     "setproperty",
	Short:   "zet de owner/permission bits of files",
	Long:    `Zet de owner/permission bits of files and directories. Only the Brocade specific names are allowed`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs setproperty process *.pdf`,
	RunE:    fsSetproperty,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsSetpropertyCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsSetpropertyCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsSetpropertyCmd)
}

func fsSetproperty(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		props := getprops()
		sort.Strings(props)
		ask = true
		fmt.Println("[", strings.Join(props, ", "), "]")
		fmt.Print("Property              : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
		if len(args) == 0 {
			return nil
		}
	}
	if len(args) == 1 {
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
	files, err := glob(Fcwd, args[1:], Frecurse, Fpattern, true, true)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qerror.ShowResult("", Fjq, err, Fyaml)
			return nil
		}
		msg := make(map[string][]string)
		msg["setproperty"] = files
		Fmsg = qerror.ShowResult(msg, Fjq, nil, Fyaml)
		return nil
	}
	pathmode := args[0]
	fn := func(n int) (interface{}, error) {
		src := files[n]
		isdir := qfs.IsDir(src)
		mode := pathmode + "file"
		if isdir {
			mode = pathmode + "dir"
		}
		err := qfs.SetPathmode(src, mode)
		return src, err
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var setproperty []string
	for i := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.setproperty"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		setproperty = append(setproperty, src)
	}

	msg := make(map[string][]string)
	msg["setproperty"] = setproperty
	if len(errs) == 0 {
		Fmsg = qerror.ShowResult(msg, Fjq, nil, Fyaml)
	} else {
		Fmsg = qerror.ShowResult(msg, Fjq, qerror.ErrorSlice(errs), Fyaml)
	}
	return nil
}

func getprops() []string {
	props := make([]string, 0)
	for key := range qregistry.Registry {
		if strings.HasPrefix(key, "fs-owner-") {
			x := strings.TrimPrefix(key, "fs-owner-")
			if x != "" {
				props = append(props, x)
			}
		}
	}
	return props
}
