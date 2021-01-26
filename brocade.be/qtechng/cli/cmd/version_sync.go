package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qsync "brocade.be/qtechng/lib/sync"
)

var versionSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Updates a production server",
	Long: `Works only on a production server. 
The command finds all changes committed to the current release
(registry value brocade-release)
and applies these changes`,
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

	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") && strings.Contains(qtechType, "P") {
		err := &qerror.QError{
			Ref: []string{"sync.bp"},
			Msg: []string{"No sync necessary: server is both production and development"},
		}
		Fmsg = qerror.ShowResult("", Fjq, err)
		return nil
	}

	current := qregistry.Registry["brocade-release"]
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"sync.production"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qerror.ShowResult("", Fjq, err)
		return nil
	}

	changed, deleted, err := qsync.Sync(current, current, false)

	if err != nil {
		Fmsg = qerror.ShowResult("", Fjq, err)
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
	Fmsg = qerror.ShowResult(msg, Fjq, nil)
	return nil
}