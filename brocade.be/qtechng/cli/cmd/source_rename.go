package cmd

import (
	"errors"
	"fmt"
	"regexp"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var sourceRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Renames sources in the repository",
	Long: `Renames sources in the repository.
The sources are specified by a combination of:

	- specific arguments
	- one or more *--qpattern* flags
	- the *--nature* flag (nature of files)
	- the *--needle* flags (text in the files)
	- the *--cuser* flags (uid of the creator)
	- the *--muser* flags (uid of the last modifier)
	- the *--cafter* flags (uid of the last modifier)

Give with the *--number* flag the number of files to be renamed.
This is a safety measure that forces the user to carefully consider
this potentially destructive command.

Without --number, no rename is performed!`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source rename --qpattern=/application/*.m --number=12`,
	RunE:    sourceRename,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
		"fill-version":      "yes",
	},
}

var Freplace string
var Fwith string
var Foverwrite bool

func init() {
	sourceRenameCmd.PersistentFlags().IntVar(&Fnumber, "number", 0, "Number of renames")
	sourceRenameCmd.PersistentFlags().StringVar(&Freplace, "replace", "", "Replace in qpath")
	sourceRenameCmd.PersistentFlags().StringVar(&Fwith, "with", "", "Replacement expression")
	sourceRenameCmd.PersistentFlags().BoolVar(&Fregexp, "regexp", false, "Replace with regular expressions")
	sourceRenameCmd.PersistentFlags().BoolVar(&Fregexp, "overwrite", false, "Allow overwrite existing sources")
	sourceCmd.AddCommand(sourceRenameCmd)
}

func sourceRename(cmd *cobra.Command, args []string) error {
	if Fwith == "" {
		err := fmt.Errorf("`--with=...` flag should not be empty")
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if Fregexp && Freplace != "" {
		_, e := regexp.Compile(Freplace)
		if e != nil {
			err := fmt.Errorf("`--replace=...` flag is not a valid regular expression")
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
	}
	squery := buildSQuery(args, Ffilesinproject, nil, false)
	froms, errs := renameData(squery, Fnumber, Freplace, Fwith, Fregexp, Foverwrite)
	if froms == nil && errs == nil {
		errs = append(errs, errors.New("no matching sources found to rename"))
	}

	type ren struct {
		Qpath string `json:"qpath"`
		From  string `json:"from"`
		To    string `json:"to"`
	}
	result := make(map[string][]ren)
	if len(froms) == 0 {
		result = nil
	} else {
		result["renames"] = make([]ren, len(froms))
		i := -1
		for fr, t := range froms {
			i++
			result["renames"][i] = ren{
				From: fr,
				To:   t,
			}
			if len(errs) == 0 {
				result["renames"][i].Qpath = t
			} else {
				result["renames"][i].Qpath = fr
			}
		}
	}
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
