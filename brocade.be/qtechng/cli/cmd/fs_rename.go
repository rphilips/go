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
	"github.com/spf13/cobra"
)

var fsRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "renames files",
	Long: `First argument is part of the absolute filepath that has to be renamed 
Second argument is the replacement of that part
The other arguments are filenames or directory names. 
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs rename cwd=../catalografie`,
	RunE:    fsRename,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsRenameCmd.Flags().BoolVar(&Fregexp, "regexp", false, "Regular expression")
	fsRenameCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recurse directories")
	fsRenameCmd.Flags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsRenameCmd)
}

func fsRename(cmd *cobra.Command, args []string) error {
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
		}
	}

	rename := args[1]
	needle := args[0]
	var rneedle *regexp.Regexp
	var err error
	files := make([]string, 0)
	if Fregexp {
		rneedle, err = regexp.Compile(needle)
	}
	if err == nil {
		files, err = glob(Fcwd, args[2:], Frecurse, Fpattern)
	}

	if len(files) == 0 {
		if err != nil {
			Fmsg = qerror.ShowResult("", Fjq, err)
			return nil
		}
		msg := make(map[string][]string)
		msg["renamed"] = files
		Fmsg = qerror.ShowResult(msg, Fjq, nil)
		return nil
	}

	renamemap := make(map[string]string)
	folders := make(map[string]string)

	target := ""

	rfiles := make([]string, 0)
	for _, file := range files {
		a, _ := filepath.Abs(file)
		if Fregexp {
			target = rneedle.ReplaceAllString(a, rename)
		} else {
			target = strings.ReplaceAll(a, needle, rename)
		}
		if target == file {
			continue
		}
		if target == a {
			continue
		}
		tdir := filepath.Dir(target)
		dir := filepath.Dir(file)
		folders[tdir] = dir
		renamemap[file] = target
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
			Fmsg = qerror.ShowResult("", Fjq, qerror.ErrorSlice(errs))
			return nil
		}

	}

	fn := func(n int) (interface{}, error) {

		src := files[n]
		dst := renamemap[src]
		err := qfs.CopyFile(src, dst, "=", true)
		if err == nil {
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
				Ref:  []string{"fs.rename"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		changed = append(changed, dst.(string))
	}

	msg := make(map[string][]string)
	msg["renamed"] = changed
	if len(errs) == 0 {
		Fmsg = qerror.ShowResult(msg, Fjq, nil)
	} else {
		Fmsg = qerror.ShowResult(msg, Fjq, qerror.ErrorSlice(errs))
	}
	return nil
}
