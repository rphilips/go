package cmd

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qinstall "brocade.be/qtechng/lib/install"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qsync "brocade.be/qtechng/lib/sync"
)

var versionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a release",
	Long: `The release matching the registry value in brocade-release
is (re)installed.

The registry value should be set with an appropriate value.`,
	Args:    cobra.NoArgs,
	Example: "qtechng version install",
	RunE:    versionInstall,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	versionInstallCmd.PersistentFlags().StringVar(&Finstallref, "installref", "", "Reference to the installation")
	versionCmd.AddCommand(versionInstallCmd)
}

func versionInstall(cmd *cobra.Command, args []string) error {

	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.version"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qerror.ShowResult("", Fjq, err)
		return nil
	}
	if Finstallref == "" {
		Finstallref = "install-" + current
	}

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
	}

	query := &qsource.Query{
		Release:  current,
		Patterns: []string{"*"},
	}

	sources := query.Run()

	err := qinstall.Install(Finstallref, sources, false)

	if err != nil {
		if err != nil {
			Fmsg = qerror.ShowResult("", Fjq, err)
			return nil
		}
	}
	msg := make(map[string][]string)
	if len(sources) != 0 {
		qpaths := make([]string, len(sources))
		for i, s := range sources {
			qpaths[i] = s.String()
		}
		sort.Strings(qpaths)
		msg["installed"] = qpaths
	}
	Fmsg = qerror.ShowResult(msg, Fjq, nil)
	return nil
}
