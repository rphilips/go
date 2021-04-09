package cmd

import (
	"sort"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsync "brocade.be/qtechng/lib/sync"
)

var versionCopyCmd = &cobra.Command{
	Use:   "copy source target",
	Short: "Copies all files from one version to another",
	Long: `The source version should always be 0.00
The target version should not exist (this condition can be removed
by using the force flag)
and should be more recent than the registry value in brocade-release`,
	Args:    cobra.ExactArgs(2),
	Example: "qtechng version copy 0.00 5.20",
	RunE:    versionCopy,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	versionCopyCmd.Flags().BoolVar(&Fforce, "force", false, "Copy even if the target version exists")
	versionCmd.AddCommand(versionCopyCmd)
}

func versionCopy(cmd *cobra.Command, args []string) error {

	rsource := args[0]
	rtarget := args[1]

	if !Fforce {
		r := qserver.Canon(rtarget)
		ro, _ := qserver.Release{}.New(r, false)
		ok, _ := ro.Exists("/source/data")
		if ok {
			err := &qerror.QError{
				Ref: []string{"copy.target.exists"},
				Msg: []string{"Target version exists. Use force!"},
			}
			Fmsg = qreport.Report("", err, Fjq, Fyaml)
			return nil
		}
	}

	changed, deleted, err := qsync.Sync(rsource, rtarget, Fforce)

	if err != nil {
		Fmsg = qreport.Report("", err, Fjq, Fyaml)
		return nil
	}
	msg := make(map[string][]string)
	if len(changed) != 0 {
		sort.Strings(changed)
		msg["copied"] = changed
	}
	if len(deleted) != 0 {
		sort.Strings(deleted)
		msg["deleted"] = deleted
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml)
	return nil
}
