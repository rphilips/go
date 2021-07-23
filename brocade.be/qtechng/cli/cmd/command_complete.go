package cmd

import (
	"strings"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"

	"github.com/spf13/cobra"
)

var commandCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Completion functionality",
	Long: `This command is a helper for providing completion information 
to Bash-like shells.

Is use is mainly for interactive`,
	Args:    cobra.ArbitraryArgs,
	Example: " qtechng command complete version",
	RunE:    commandComplete}

func init() {

	commandCmd.AddCommand(commandCompleteCmd)
}

func commandComplete(cmdo *cobra.Command, args []string) error {

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
	result := make([]string, 0)
	if len(argums) == 0 {
		for _, cmd := range rootCmd.Commands() {
			if checkQtCmd(cmd, QtechType) {
				parts := strings.SplitN(cmd.Use, " ", 2)
				result = append(result, parts[0])
			}

		}
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	verb := argums[0]
	cmd := searchCmd(verb, rootCmd.Commands())
	if cmd == nil {
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	cmds := cmd.Commands()
	if !checkQtCmd(cmd, QtechType) {
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if len(cmds) == 0 {
		result = getComplete(cmd, argums[1:], QtechType)
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if len(argums) == 1 {
		for _, cm := range cmds {
			if checkQtCmd(cm, QtechType) {
				parts := strings.SplitN(cm.Use, " ", 2)
				result = append(result, parts[0])
			}
		}
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	cmd = searchCmd(argums[1], cmds)
	result = getComplete(cmd, argums[2:], QtechType)

	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil

}

func getComplete(cmd *cobra.Command, used []string, qt string) []string {
	result := make([]string, 0)
	qtcomplete := cmd.Annotations["complete"]
	if qtcomplete == "" {
		return result
	}
	switch qtcomplete {
	case "version":
		if strings.Contains(qt, "P") {
			return []string{qregistry.Registry["brocade-release"]}
		}
		if strings.Contains(qt, "B") {
			v := qserver.Releases(0)
			vs := strings.SplitN(v, " ", -1)
			for _, v := range vs {
				ok := false
				for _, v2 := range used {
					if v2 == v {
						ok = true
						break
					}
				}
				if !ok {
					result = append(result, v)
				}
			}
			return result
		}
		if strings.Contains(qt, "W") {
			v := qregistry.Registry["qtechng-releases"]
			vs := strings.SplitN(v, " ", -1)
			for _, v := range vs {
				ok := false
				for _, v2 := range used {
					if v2 == v {
						ok = true
						break
					}
				}
				if !ok {
					result = append(result, v)
				}
			}
			return result
		}
	}

	return result
}

func searchCmd(name string, cmds []*cobra.Command) *cobra.Command {
	for _, cmd := range cmds {
		parts := strings.SplitN(cmd.Use, " ", 2)
		if parts[0] != name {
			continue
		}
		return cmd
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
