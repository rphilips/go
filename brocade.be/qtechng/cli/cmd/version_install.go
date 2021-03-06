package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
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
	versionInstallCmd.PersistentFlags().StringVar(&Frefname, "refname", "", "Reference to the installation")
	versionCmd.AddCommand(versionInstallCmd)
}

func versionInstall(cmd *cobra.Command, args []string) error {

	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.version"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if Frefname == "" {
		Frefname = "install-" + current
	}

	logme := log.New(os.Stderr, Frefname+" ", log.LstdFlags)
	t0 := time.Now()

	logme.Println("Start")

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
		logme.Println(fmt.Sprintf("Synchronised version `%s` with dev.anet.be", current))
	}

	query := &qsource.Query{
		Release:  current,
		Patterns: []string{"*"},
	}

	sources := query.Run()

	err := qsource.Install(Frefname, sources, false, logme)
	reportfile := filepath.Join(qregistry.Registry["scratch-dir"], Frefname+".json")
	t1 := time.Now()
	logme.Printf("Results also in `%s`", reportfile)
	logme.Printf("Runtime: %v\n", t1.Sub(t0))
	logme.Println("End")

	if err != nil {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, reportfile)
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
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, reportfile)
	return nil
}
