package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qsource "brocade.be/qtechng/lib/source"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
)

var versionSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize a production server",
	Long: `This command synchronizes the current release of a production server (*brocade-release*)
with the most recent version of that release on the development server.
The command finds all changes committed to the current release
(registry value brocade-release) and applies these changes.
Works only on a production server.
Depending on the underlying instruction (registry value qtechng-sync-exe)
it may be necessary to run this command as root!`,
	Args:    cobra.NoArgs,
	Example: "qtechng version sync",
	RunE:    versionSync,
	Annotations: map[string]string{
		"with-qtechtype": "P",
	},
}

func init() {
	versionCmd.AddCommand(versionSyncCmd)
}

func versionSync(cmd *cobra.Command, args []string) error {
	// if runtime.GOOS == "linux" {
	// 	user, _ := user.Current()
	// 	if user.Username != "root" {
	// 		fmt.Println("\nTry:\nsudo qtechng version sync")
	// 		return nil
	// 	}
	// }

	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") && strings.Contains(qtechType, "P") {
		err := &qerror.QError{
			Ref: []string{"sync.bp"},
			Msg: []string{"No sync necessary: server is both production and development"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	current := qregistry.Registry["brocade-release"]
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"sync.production"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	changed, deleted, err := qsync.Sync(current, current, false)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	msg := make(map[string][]string)
	if len(changed) != 0 {
		sort.Strings(changed)
		msg["synced"] = changed
	}
	if len(deleted) != 0 {
		sort.Strings(deleted)
		msg["deleted"] = deleted
	}

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	work := make([]string, 0)
	work = append(work, changed...)
	work = append(work, deleted...)

	query := &qsource.Query{
		Release:  current,
		Patterns: work,
	}
	sources := query.Run()
	refname := qutil.Reference("synced")
	err = qsource.Install(refname, sources, false, nil)

	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
