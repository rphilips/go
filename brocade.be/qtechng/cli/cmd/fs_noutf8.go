package cmd

import (
	"bufio"
	"errors"
	"os"
	"path"
	"sort"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsNoutf8Cmd = &cobra.Command{
	Use:   "noutf8",
	Short: "Search for non-UTF8 sequences",
	Long: `This command searches for non-UTF8 byte sequences

	The arguments are files or directories.
	A directory stand for ALL its files.

	These argument scan be expanded/restricted by using the flags:

		- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
		- The '--pattern' flag builds a list of acceptable patterns on the basenames
		- The '--utf8only' flag restricts to files with UTF-8 content


	Some remarks:

		- Search is done line per line
		- With the '--ask' flag, you can interactively specify the arguments and flags
		- The result is with file URLs (appropriate for fast editor location)`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs noutf8 *.m cwd=../catalografie`,
	RunE:    fsNoutf8,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsNoutf8Cmd)
}

func fsNoutf8(cmd *cobra.Command, args []string) error {
	Futf8only = true
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-noutf8-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
	}

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, false)
		if err != nil {
			Ferrid = "fs-noutf8-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-noutf8-nofiles")
		return nil
	}

	fn := func(n int) (interface{}, error) {
		lines := make([]int, 0)
		src := files[n]
		in, err := os.Open(src)
		if err != nil {
			return lines, err
		}
		input := bufio.NewReader(in)
		defer in.Close()

		_, problems, _ := qutil.NoUTF8(input)

		for _, problem := range problems {
			lineno := problem[0]
			if lineno > 0 {
				lines = append(lines, lineno)
			}
		}
		return lines, nil
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	type rtt struct {
		All []string            `json:"all"`
		Ext map[string][]string `json:"ext"`
	}
	rt := rtt{}
	rt.All = make([]string, 0)
	rt.Ext = make(map[string][]string)
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.noutf8"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		linenos := r.([]int)
		if len(linenos) == 0 {
			continue
		}
		ext := path.Ext(src)

		lst := rt.Ext[ext]
		if len(lst) == 0 {
			lst = make([]string, 0)
		}
		for _, lineno := range linenos {
			rt.All = append(rt.All, qutil.FileURL(src, "", lineno))
			lst = append(lst, qutil.FileURL(src, "", lineno))
		}
		rt.Ext[ext] = lst
	}
	if len(errs) == 0 {
		errs = nil
	}
	lst := rt.All
	sort.Strings(lst)
	rt.All = lst

	for ext, lst := range rt.Ext {
		sort.Strings(lst)
		rt.Ext[ext] = lst
	}

	Fmsg = qreport.Report(rt, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
