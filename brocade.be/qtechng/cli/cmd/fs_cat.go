package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsCatCmd = &cobra.Command{
	Use:     "cat",
	Short:   "cat a file",
	Long:    `First argument is part of the absolute filepath that has to be copied to stdout`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng fs cat bcawedit.m cwd=../catalografie`,
	RunE:    fsCat,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsCatCmd)
}

func fsCat(cmd *cobra.Command, args []string) error {
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
		if len(args) != 0 {
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
	var files []string
	var err error

	if len(args) != 1 || args[0] != "-" {
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false)
		if len(files) == 0 {
			if err != nil {
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
				return nil
			}
			msg := make(map[string][]string)
			msg["copied"] = files
			Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote)
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
	if len(args) == 1 || args[0] == "-" {
		io.Copy(output, os.Stdin)
	}
	return nil
}
