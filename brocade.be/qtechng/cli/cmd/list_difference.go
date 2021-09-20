package cmd

import (
	"errors"
	"strconv"

	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var listDifferenceCmd = &cobra.Command{
	Use:   "difference",
	Short: "Difference of QtechNG lists",
	Long: `Constructs a new list out of the difference of existing QtechNG lists
(see lists folder in qtechng-support-dir)

The result is kept in a list identified by the value of the '--list=...' flag.
Note that the contents of this (new) list will be overwritten.

`,
	Args:    cobra.MinimumNArgs(2),
	Example: `qtechng list difference mylist1 mylist2 --list=elementinlist1andnotinlist2`,
	RunE:    listDifference,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

func init() {
	listCmd.AddCommand(listDifferenceCmd)
}

func listDifference(cmd *cobra.Command, args []string) error {
	if Flist == "" {
		return errors.New("--list=... flag is missing")
	}
	if !qutil.ListTest(Flist) {
		return errors.New("list should start with lowercase letter followed by lowercase letters, numbers or underscore")
	}
	list1 := qutil.GetLists(args[0:1])
	list2 := qutil.GetLists(args[1:])
	result1 := make(map[string]bool)
	result2 := make(map[string]bool)

	for _, one := range list1 {
		for k := range one {
			result1[k] = true
		}
	}

	for _, one := range list2 {
		for k := range one {
			result2[k] = true
		}
	}

	iresult := make([]string, 0)

	for qpath := range result1 {
		if result2[qpath] {
			continue
		}
		iresult = append(iresult, qpath)
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
