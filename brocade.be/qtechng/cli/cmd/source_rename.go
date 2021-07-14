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
- by one or more *--qpattern* flags
- by the specification of the nature of the files with the *--nature* flag
- by specification of *--needle* flags (text in the files)
- by specification of *--cuser* flags (uid of the creator)
- by specification of *--muser* flags (uid of the last modifier)
- by specification of *--cafter* flags (uid of the last modifier)

Give with the *--number* flag the number of files to be deleted.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source rename --qpattern=/application/*.m --number=12`,
	RunE:    sourceRename,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
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
	sourceRenameCmd.PersistentFlags().IntVar(&Fnumber, "number", 0, "number of deletes")
	sourceRenameCmd.PersistentFlags().StringVar(&Freplace, "replace", "", "replace in qpath")
	sourceRenameCmd.PersistentFlags().StringVar(&Fwith, "with", "", "replacement expression")
	sourceRenameCmd.PersistentFlags().BoolVar(&Fregexp, "regexp", false, "replace with regular expressions")
	sourceRenameCmd.PersistentFlags().BoolVar(&Fregexp, "overwrite", false, "allow overwrite existing sources")
	sourceCmd.AddCommand(sourceRenameCmd)
}

func sourceRename(cmd *cobra.Command, args []string) error {
	if Fwith == "" {
		err := fmt.Errorf("`--with=...` flag should not be empty")
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if Fregexp && Freplace != "" {
		_, e := regexp.Compile(Freplace)
		if e != nil {
			err := fmt.Errorf("`--replace=...` flag is not a valid regular expression")
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
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
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
