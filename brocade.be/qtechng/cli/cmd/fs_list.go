package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	Long: `The last modification time of all files is changed to the current moment.
If the argument is a directory name, all files in that directory are handled.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs list cwd=../catalografie`,
	RunE:    fsList,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Fonlytext bool

func init() {
	fsListCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsListCmd.Flags().BoolVar(&Fonlytext, "onlytext", false, "Only text files")
	fsListCmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsListCmd)
}

func fsList(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	ask := false
	if len(args) == 0 {
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
			if text == "*" {
				break
			}
		}
	}

	if ask && !Fonlytext {
		fmt.Print("Only text files ?       : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Fonlytext = true
		}
	}

	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false, Fonlytext)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	msg := make(map[string][]string)
	msg["listed"] = files
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
