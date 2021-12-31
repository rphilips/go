package cmd

import (
	"errors"
	"io"
	"os"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsCatCmd = &cobra.Command{
	Use:   "cat",
	Short: "Execute the cat command ",
	Long: `Execute  the cat command

	Some remarks:

	- With no arguments, stdin is copied
	  output is written to stdout
	- the arguments are files or directories
	  (use '.' to indicate the current working directory)
	- If an argument is a directory, all files in that directory are handled.
	- With the '--ask' flag, you can interactively specify the arguments and flags
	- The '--recurse' flag recursively traverses the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content
    - With the '--stdout=...' flag, the contents can be redirected to a file.`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs cat bcawedit.m cwd=../workspace
qtechng fs cat . cwd=../workspace --stdout=result.txt
qtechng fs cat --ask
`,
	RunE: fsCat,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsCatCmd)
}

func fsCat(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files:",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-cat-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}

	var files []string
	var err error

	if len(args) != 0 {
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, false)
		if err != nil {
			Ferrid = "fs-cat-glob"
			return err
		}
		if len(files) == 0 {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-cat-nofiles")
			return nil
		}
	}

	output := os.Stdout
	if Fstdout != "" {
		f, err := os.Create(qutil.AbsPath(Fstdout, Fcwd))
		if err != nil {
			return err
		}
		output = f
		defer output.Close()
	}
	for _, file := range files {
		f, err := os.Open(qutil.AbsPath(file, Fcwd))
		if err != nil {
			continue
		}
		io.Copy(output, f)
		f.Close()
	}
	if len(args) == 0 {
		io.Copy(output, os.Stdin)
	}
	return nil
}
