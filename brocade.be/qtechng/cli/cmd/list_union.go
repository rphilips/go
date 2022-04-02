package cmd

import (
	"errors"
	"strconv"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var listUnionCmd = &cobra.Command{
	Use:   "union",
	Short: "Union of QtechNG lists",
	Long: `Constructs a new list out of the union of existing QtechNG lists
(see lists folder in qtechng-support-dir).

The result is kept in a list identified by the value of the '--list=...' flag.
Note that the contents of this (new) list will be overwritten.

The 'union' can also be used to copy lists.
`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng list union mylist1 mylist2 --list=unionof1and2`,
	RunE:    listUnion,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

func init() {
	listCmd.AddCommand(listUnionCmd)
}

func listUnion(cmd *cobra.Command, args []string) error {
	if Flist == "" {
		return errors.New("--list=... flag is missing")
	}
	if !qutil.ListTest(Flist) {
		return errors.New("list should start with lowercase letter followed by lowercase letters, numbers or underscore")
	}
	lists := qutil.GetLists(args)
	result := make(map[string]bool)

	for _, one := range lists {
		for k := range one {
			result[k] = true
		}
	}
	lresult := make([]string, len(result))
	i := 0
	for qpath := range result {
		lresult[i] = qpath
		i++
	}

	if len(lresult) != 0 {
		qutil.EditList(Flist, false, lresult)
		if Fshow {
			Fmsg = qreport.Report(lresult, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		} else {
			Fmsg = qreport.Report("Created `"+Flist+"` with "+strconv.Itoa(len(lresult))+" elements", nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
	} else {
		Fmsg = qreport.Report("No elements found!", nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}

	return nil
}
