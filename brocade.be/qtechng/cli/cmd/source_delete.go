package cmd

import (
	"errors"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var sourceDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete sources in the repository",
	Long: `This command deletes sources in the repository.
The sources are specified by a combination of:

	- specific arguments
	- one or more *--qpattern* flags
	- the *--nature* flag (nature of files)
	- the *--needle* flags (text in the files)
	- the *--cuser* flags (uid of the creator)
	- the *--muser* flags (uid of the last modifier)
	- the *--cafter* flags (uid of the last modifier)

Use the --number flag to indicate the number of files to be deleted.
This is a safety measure that forces the user to carefully consider
this destructive command.

Without --number, no delete is performed!`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source delete --qpattern=/application/*.m --number=12`,
	RunE:    sourceDelete,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
		"fill-version":      "yes",
	},
}

var Fnumber int

func init() {
	sourceDeleteCmd.PersistentFlags().IntVar(&Fnumber, "number", 0, "Number of deletes")
	sourceCmd.AddCommand(sourceDeleteCmd)
}

func sourceDelete(cmd *cobra.Command, args []string) error {
	squery := buildSQuery(args, Ffilesinproject, nil, false)
	qpaths, errs := delData(squery, Fnumber)
	if qpaths == nil && errs == nil {
		errs = errors.New("no matching sources found to delete")
	}
	result := make(map[string][]string)
	if len(qpaths) == 0 {
		result = nil
	} else {
		result["qpath"] = qpaths
	}
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
