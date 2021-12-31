package cmd

import (
	"errors"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var Fproperty = ""

var fsSetpropertyCmd = &cobra.Command{
	Use:   "setproperty",
	Short: "Set owner/permission bits of files",
	Long: `This command sets the owner/permission bits of files and directories.

Only the Brocade specific names are allowed:
	- naked
	- process
	- qtech
	- script
	- temp
	- web
	- webdav

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

Some remarks:

	- Search is done line per line
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- With the '--property' flag, you can specify the suitable property`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs setproperty *.pdf --property=process
qtechng fs setproperty --ask`,
	RunE: fsSetproperty,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsSetpropertyCmd.Flags().StringVar(&Fproperty, "property", "", "Brocade property")
	fsCmd.AddCommand(fsSetpropertyCmd)
}

func fsSetproperty(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"property:files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-setproperty-abort")
			return nil
		}
		args = argums["files"].([]string)
		Fproperty = argums["property"].(string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, true, Futf8only)
		if err != nil {
			Ferrid = "fs-setproperty-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-setproperty-nofiles")
		return nil
	}

	pathmode := Fproperty
	props := getprops()
	if !props[pathmode] {
		Fmsg = qreport.Report(nil, errors.New("invalid property `"+pathmode+"`"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-setproperty-invalid")
		return nil
	}

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
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-setproperty-notset")
	}
	return nil
}

func getprops() (props map[string]bool) {
	props = make(map[string]bool)
	for key := range qregistry.Registry {
		if strings.HasPrefix(key, "fs-owner-") {
			x := strings.TrimPrefix(key, "fs-owner-")
			if x != "" {
				props[x] = true
			}
		}
	}
	return props
}
