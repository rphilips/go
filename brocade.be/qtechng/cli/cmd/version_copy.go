package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
)

var versionCopyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Retrieves a version form the development server",
	Long: `Works only on a production server. 
The command finds all changes committed to the given release
(registry value brocade-release)
and applies these changes`,
	Args:    cobra.RangeArgs(1, 2),
	Example: "qtechng version copy 5.40\nqtechng version copy 0.00 5.50",
	RunE:    versionCopy,
	Annotations: map[string]string{
		"with-qtechtype": "P",
	},
}

func init() {
	versionCmd.AddCommand(versionCopyCmd)
}

func versionCopy(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		args = append(args, args[0])
	}
	tversion := args[1]
	sversion := args[0]
	if sversion != tversion && sversion != "0.00" {
		err := &qerror.QError{
			Ref: []string{"copy.unequal"},
			Msg: []string{"Only `0.00` can be copied to a different version"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	qtechType := qregistry.Registry["qtechng-type"]
	if strings.ContainsRune(qtechType, 'B') && strings.ContainsRune(qtechType, 'P') {
		err := &qerror.QError{
			Ref: []string{"copy.bp"},
			Msg: []string{"No copy necessary: server is both production and development"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	current := qregistry.Registry["brocade-release"]
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"copy.production"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	if tversion == "0.00" {
		err := &qerror.QError{
			Ref: []string{"copy.dev.to000"},
			Msg: []string{"cannot copy to version `0.00`"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	if tversion == current {
		err := &qerror.QError{
			Ref: []string{"copy.current"},
			Msg: []string{"Cannot copy the current release: use `qtechng version sync`"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	lowest := qutil.LowestVersion(current, tversion)
	if current != lowest {
		err := &qerror.QError{
			Ref: []string{"copy.version.production.lowest"},
			Msg: []string{"The version should be higher than the current version"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil

	}

	changed, deleted, err := qsync.Sync(sversion, tversion, false)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	msg := make(map[string][]string)
	if len(changed) != 0 {
		sort.Strings(changed)
		msg["copyed"] = changed
	}
	if len(deleted) != 0 {
		sort.Strings(deleted)
		msg["deleted"] = deleted
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
