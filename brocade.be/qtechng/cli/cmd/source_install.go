package cmd

import (
	"sort"
	"strings"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install sources in the repository",
	Long:    `This command installs sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source install --qpattern=/application/*.m`,
	RunE:    sourceInstall,
	PreRun:  preSourceInstall,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceInstallCmd.PersistentFlags().StringVar(&Frefname, "refname", "install", "Reference to the installation")
	sourceInstallCmd.Flags().BoolVar(&Fwarnings, "warnings", false, "Include warnings")
	sourceCmd.AddCommand(sourceInstallCmd)
}

func sourceInstall(cmd *cobra.Command, args []string) error {

	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.source"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if Frefname == "" {
		Frefname = "sourceinstall-" + qutil.Timestamp(true)
	}

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true, false, false)
	}

	patterns := make([]string, len(args))
	copy(patterns, args)

	query := &qsource.Query{
		Release:  current,
		Patterns: patterns,
	}

	sources := query.Run()

	err := qsource.Install(Frefname, sources, Fwarnings, nil)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
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
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}

func preSourceInstall(cmd *cobra.Command, args []string) {
	if Frefname == "" {
		Frefname = "sourceinstall-" + qutil.Timestamp(true)
	}
	if strings.Contains(QtechType, "P") {
		qsync.Sync("", "", true, false, false)
	}

	if !strings.ContainsAny(QtechType, "BP") {
		preSSH(cmd, nil)
	}

}
