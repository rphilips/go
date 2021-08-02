package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "copys files",
	Long: `First argument is part of the absolute filepath that has to be copied
Second argument is the replacement of that part
The other arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.
Use the delete flag if the original files should be deleted
`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs copy cwd=../catalografie`,
	RunE:    fsCopy,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

//Fdelete delete after copy
var Fdelete bool

// Fconfirm ask for confirmation
var Fconfirm bool

func init() {
	fsCopyCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsCopyCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsCopyCmd.Flags().BoolVar(&Fdelete, "delete", false, "Delete original files")
	fsCopyCmd.Flags().BoolVar(&Fconfirm, "confirm", false, "Ask the first time for confirmation")
	fsCopyCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsCopyCmd)
}

func fsCopy(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false
	if len(args) == 0 {
		ask = true
		fmt.Print("Enter search string in file name: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			return nil
		}
		args = append(args, text)
	}
	if len(args) == 1 {
		ask = true
		fmt.Print("Enter replacement string   : ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		args = append(args, text)
	}

	if len(args) == 2 {
		ask = true
		for {
			fmt.Print("File/directory             : ")
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
	if ask && !Fdelete {
		fmt.Print("Delete ?                   : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fdelete = true
		}
	}
	if ask && !Fregexp {
		fmt.Print("Regexp ?                   : <n>")
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
		fmt.Print("Recurse ?                  : <n>")
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
			fmt.Print("Pattern on basename        : ")
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

	if ask && !Fconfirm {
		fmt.Print("Confirm first time ?       : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fconfirm = true
		}
	}

	copy := args[1]
	needle := args[0]
	var rneedle *regexp.Regexp
	var err error
	files := make([]string, 0)
	if Fregexp {
		rneedle, err = regexp.Compile(needle)
	}
	if err == nil {
		files, err = glob(Fcwd, args[2:], Frecurse, Fpattern, true, false, false)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["copied"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	copymap := make(map[string]string)
	folders := make(map[string]string)

	target := ""

	rfiles := make([]string, 0)
	for _, file := range files {
		a, _ := filepath.Abs(file)
		if Fregexp {
			target = rneedle.ReplaceAllString(a, copy)
		} else {
			target = strings.ReplaceAll(a, needle, copy)
		}
		if target == file {
			continue
		}
		if target == a {
			continue
		}
		if Fconfirm {
			Fconfirm = false
			fmt.Printf("Copy `%s` to `%s`: <n>", a, target)
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
				rfiles = make([]string, 0)
				folders = make(map[string]string)
				break
			}
		}
		tdir := filepath.Dir(target)
		dir := filepath.Dir(file)
		folders[tdir] = dir
		copymap[file] = target
		rfiles = append(rfiles, file)
	}

	dirs := make([]string, len(folders))

	for dir := range folders {
		dirs = append(dirs, dir)
	}

	sort.Strings(dirs)

	errs := make([]error, 0)
	for _, dir := range dirs {
		if qfs.IsDir(dir) {
			continue
		}
		if qfs.Exists(dir) {
			errs = append(errs, fmt.Errorf("`%s` exists and is not a directory", dir))
		}
		err := qfs.Mkdir(dir, "process")
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = qfs.CopyMeta(folders[dir], dir, true)
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) != 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}

	}

	fn := func(n int) (interface{}, error) {

		src := files[n]
		dst := copymap[src]
		err := qfs.CopyFile(src, dst, "=", true)
		if err == nil && Fdelete {
			err = qfs.Rmpath(src)
		}
		return dst, err
	}

	resultlist, errorlist := qparallel.NMap(len(rfiles), -1, fn)
	var changed []string
	for i, dst := range resultlist {
		src := rfiles[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.copy"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		changed = append(changed, dst.(string))
	}

	msg := make(map[string][]string)
	msg["copied"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
