package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy/Rename files",
	Long: `This action allows for copying and renaming of files.

Warning! This command is very powerful and can permanently alter your files.

The arguments are files or directories.
A directory stand for ALL its files.



These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

The target destination is constructed by a search/replace action on the arguments:

    - The '--search' flag gives the string to search for in the abspath of the argument
    - The '--replace' flag replaces the 'search' part
	- The '--regexp' flag indicates if the '--search' flag is a regular expression

There are 3 special flags to add functionality:

    - With the '--ask' flag, you can interactively specify the arguments and flags
	- With the '--confirm' flag, you can inspect the FIRST replacement BEFORE it is
	  executed.
	- With the '--delete' flag, you can delete the original file`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs copy . --recurse --pattern='*.jpeg' --search=.jpeg --replace=.jpg --cwd=../workspace
qtechng fs copy --ask`,
	RunE: fsCopy,
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
	fsCopyCmd.Flags().StringVar(&Fsearch, "search", "", "Search for")
	fsCopyCmd.Flags().StringVar(&Freplace, "replace", "", "Replace with")
	fsCopyCmd.Flags().BoolVar(&Fdelete, "delete", false, "Delete original files")
	fsCopyCmd.Flags().BoolVar(&Fconfirm, "confirm", false, "Ask the first time for confirmation")
	fsCmd.AddCommand(fsCopyCmd)
}

func fsCopy(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"search::" + Fsearch,
			"regexp:search:" + qutil.UnYes(Fregexp),
			"replace:search:" + Freplace,
			"files:search",
			"recurse:search,files:" + qutil.UnYes(Frecurse),
			"patterns:search,files:",
			"utf8only:search,files:" + qutil.UnYes(Futf8only),
			"confirm:search,files:" + qutil.UnYes(Fconfirm),
			"delete:search,files:" + qutil.UnYes(Fdelete),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-abort")
			return nil
		}
		Fsearch = argums["search"].(string)
		Freplace = argums["replace"].(string)
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fregexp = argums["regexp"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fconfirm = argums["confirm"].(bool)
		Fdelete = argums["delete"].(bool)
	}

	if Fsearch == "" {
		Fmsg = qreport.Report(nil, errors.New("search string is empty"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-search")
		return nil
	}
	var rsearch *regexp.Regexp
	if Fregexp {
		var err error
		rsearch, err = regexp.Compile(Fsearch)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-invalidregexp")
			return nil
		}
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-copy-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no files found to copy"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-nofiles")
		return nil
	}

	copymap := make(map[string]string)
	folders := make(map[string]string)

	target := ""

	rfiles := make([]string, 0)
	for _, file := range files {
		a, _ := filepath.Abs(file)
		if Fregexp {
			target = rsearch.ReplaceAllString(a, Freplace)
		} else {
			target = strings.ReplaceAll(a, Fsearch, Freplace)
		}
		if qfs.SameFile(target, a) {
			continue
		}
		if Fconfirm {
			Fconfirm = false
			prompt := fmt.Sprintf("Copy `%s` to `%s` ? ", a, target)
			confirm := qutil.Confirm(prompt)
			if !confirm {
				Fmsg = qreport.Report(nil, errors.New("did not pass confirmation"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-confirm")
				return nil
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
		Fmsg = qreport.Report(nil, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-copy-dirconstruction")
		return nil
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
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
