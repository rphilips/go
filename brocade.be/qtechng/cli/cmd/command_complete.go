package cmd

import (
	"strings"

	qreport "brocade.be/qtechng/lib/report"

	"github.com/spf13/cobra"
)

var commandCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Completion functionality",
	Long: `
Displays a list of all available qtechng commands`,
	Args:    cobra.ArbitraryArgs,
	Example: "  qtechng command complete version",
	RunE:    commandComplete}

func init() {

	commandCmd.AddCommand(commandCompleteCmd)
}

func commandComplete(cmd *cobra.Command, args []string) error {

	argums := make([]string, 0)

	for i, arg := range args {
		if i == len(args)-1 {
			continue
		}
		if i == 0 && (strings.HasSuffix(arg, "qtechng") || strings.HasSuffix(arg, "qtechng.exe")) {
			continue
		}
		if arg == "" {
			continue
		}
		argums = append(argums, arg)
	}

	if len(argums) == 0 {
		result := make([]string, 0)
		for _, cmd := range rootCmd.Commands() {
			if checkQtCmd(cmd, QtechType) {
				parts := strings.SplitN(cmd.Use, " ", 2)
				result = append(result, parts[0])
			}

		}
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}
	if len(argums) == 1 {
		result := make([]string, 0)
		verb := argums[0]
		for _, cmd := range rootCmd.Commands() {
			parts := strings.SplitN(cmd.Use, " ", 2)
			if parts[0] != verb {
				continue
			}
			if !checkQtCmd(cmd, QtechType) {
				Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fsilent)
				return nil
			}
			scmd := cmd.Commands()
			if len(scmd) == 0 {
				Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fsilent)
				return nil
			}
			for _, cm := range scmd {
				if checkQtCmd(cm, QtechType) {
					parts := strings.SplitN(cm.Use, " ", 2)
					result = append(result, parts[0])
				}
			}
			Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fsilent)
			return nil
		}
	}
	return nil
}

func checkQtCmd(cmd *cobra.Command, qt string) bool {
	if !checkQtAnno(cmd, qt) {
		return false
	}
	scmd := cmd.Commands()
	if len(scmd) == 0 {
		return true
	}
	for _, cm := range scmd {
		if checkQtCmd(cm, qt) {
			return true
		}
	}
	return false
}

func checkQtAnno(cmd *cobra.Command, qt string) bool {
	if cmd.Annotations == nil {
		return true
	}
	qtanno, ok := cmd.Annotations["with-qtechtype"]
	if !ok {
		return true
	}
	for _, r := range qt {
		if strings.ContainsRune(qtanno, r) {
			return true
		}
	}
	return false
}
