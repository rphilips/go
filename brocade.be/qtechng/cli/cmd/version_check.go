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
	qutil "brocade.be/qtechng/lib/util"
)

var versionCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check a release",
	Long: `This command executes check.py throughout all projects of the current version.

The registry value should be set with an appropriate value (*qtechng version set*).`,
	Args:    cobra.NoArgs,
	Example: "qtechng version check",
	RunE:    versionCheck,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	versionCheckCmd.PersistentFlags().StringVar(&Frefname, "refname", "", "Reference to the check")
	versionCmd.AddCommand(versionCheckCmd)
}

func versionCheck(cmd *cobra.Command, args []string) error {

	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"check.version"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if Frefname == "" {
		Frefname = "check-" + current
	}
	Frefname = qutil.Reference(Frefname)

	logme := log.New(os.Stderr, Frefname+" ", log.LstdFlags)
	t0 := time.Now()

	logme.Println("Start")

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
		logme.Println(fmt.Sprintf("Synchronised version `%s` with dev.anet.be", current))
	}

	query := &qsource.Query{
		Release:  current,
		Patterns: []string{"*/check.py"},
	}

	sources := query.Run()

	err := qsource.Check(Frefname, sources, Fwarnings, logme)
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
		msg["checked"] = qpaths
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, reportfile)
	return nil
}