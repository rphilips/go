package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
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
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version copy 5.40",
	RunE:    versionCopy,
	Annotations: map[string]string{
		"with-qtechtype": "P",
	},
}

func init() {
	versionCmd.AddCommand(versionCopyCmd)
}

func versionCopy(cmd *cobra.Command, args []string) error {

	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") && strings.Contains(qtechType, "P") {
		err := &qerror.QError{
			Ref: []string{"copy.bp"},
			Msg: []string{"No copy necessary: server is both production and development"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
		return nil
	}

	current := qregistry.Registry["brocade-release"]
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"copy.production"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
		return nil
	}

	if args[0] == "0.00" {
		err := &qerror.QError{
			Ref: []string{"copy.open"},
			Msg: []string{"Cannot copy version `0.00`"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
		return nil
	}

	version := qserver.Canon(args[0])

	if version == current {
		err := &qerror.QError{
			Ref: []string{"copy.current"},
			Msg: []string{"Cannot copy the current release: use `qtechng system sync`"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
		return nil
	}

	lowest := qutil.LowestVersion(current, version)
	if current != lowest {
		err := &qerror.QError{
			Ref: []string{"copy.version.production.lowest"},
			Msg: []string{"The version should be higher than the current version"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
		return nil

	}

	changed, deleted, err := qsync.Sync(version, version, false)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote)
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
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote)
	return nil
}
