package cmd

import (
	"errors"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	Long: `List files

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

Some remarks:

	- Search is done line per line
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- With the '--url' flag, files are shown as URLs`,

	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs list . --cwd=../workspace
	qtechng fs list --ask
`,
	RunE: fsList,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Fasurl = false

func init() {
	fsListCmd.Flags().BoolVar(&Fasurl, "url", false, "Show as URL")
	fsCmd.AddCommand(fsListCmd)
}

func fsList(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"url:files:" + qutil.UnYes(Fasurl),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-list-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Fasurl = argums["url"].(bool)
		Futf8only = argums["utf8only"].(bool)
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-list-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-list-nofiles")
		return nil
	}

	msg := make(map[string][]string)
	if Fasurl {
		for i, file := range files {
			files[i] = qutil.FileURL(file, "", -1)
		}
	}
	msg["listed"] = files
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
