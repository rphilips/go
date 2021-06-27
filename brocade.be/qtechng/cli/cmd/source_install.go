package cmd

import (
	"log"
	"sort"
	"strings"

	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
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
	Short:   "Installs sources in the repository",
	Long:    `Installs sources in the repository according to patterns, nature and contents`,
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
	sourceCmd.AddCommand(sourceInstallCmd)
}

func sourceInstall(cmd *cobra.Command, args []string) error {

	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.source"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if Frefname == "" {
		Frefname = "sourceinstall-" + qutil.Timestamp(true)
	}

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
	}

	patterns := make([]string, len(args))
	copy(patterns, args)

	query := &qsource.Query{
		Release:  current,
		Patterns: patterns,
	}

	sources := query.Run()

	err := qsource.Install(Frefname, sources, true, nil)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
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
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}

func preSourceInstall(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_install/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		installData(Fpayload, Fcargo, false, true, "", nil)
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_install/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
