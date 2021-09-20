package cmd

import (
	"errors"
	"strconv"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var listIntersectionCmd = &cobra.Command{
	Use:   "intersection",
	Short: "Intersection of QtechNG lists",
	Long: `Constructs a new list out of the intersection of existing Qtechng lists.
(see lists folder in qtechng-support-dir).

The result is kept in a list identified by the value of the '--list=...' flag.
Note that the contents of this (new) list will be overwritten.
`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng list intersection mylist1 mylist2 --list=instersectionof1and2`,
	RunE:    listIntersection,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

func init() {
	listCmd.AddCommand(listIntersectionCmd)
}

func listIntersection(cmd *cobra.Command, args []string) error {
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
	iresult := make([]string, 0)

	for qpath := range result {
		ok := true
		for _, one := range lists {
			_, ok = one[qpath]
			if !ok {
				break
			}
		}
		if ok {
			iresult = append(iresult, qpath)
		}
	}
	if len(iresult) != 0 {
		qutil.EditList(Flist, false, iresult)
		if Fshow {
			Fmsg = qreport.Report(iresult, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		} else {
			Fmsg = qreport.Report("Created `"+Flist+"` with "+strconv.Itoa(len(iresult))+" elements", nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		}
	} else {
		Fmsg = qreport.Report("No elements found!", nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}
