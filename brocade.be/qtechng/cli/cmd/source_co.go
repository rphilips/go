package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceCoCmd = &cobra.Command{
	Use:   "co",
	Short: "Check out qtechng source files",
	Long: `This command retrieve source files from the qtechng repository.
The --copyonly flag updates the local file contents, but does not affect its qtechng status.
This can be used, for instance, to deliberately replace
one repositority version of a file with another.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source co --qpattern=/catalografie/application/bcawedit.m`,
	RunE:    sourceCo,
	PreRun:  preSourceCo,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

// Fclear Clears visited directories, if in auto mode
var Fclear bool
var Fcopyonly bool

func init() {
	sourceCoCmd.Flags().BoolVar(&Fclear, "clear", false, "Clear visited directories, if in auto mode")
	sourceCoCmd.Flags().BoolVar(&Fcopyonly, "copyonly", false, "Check out without updating local repository information")
	sourceCmd.AddCommand(sourceCoCmd)
}

func sourceCo(cmd *cobra.Command, args []string) error {
	qdir := ""
	if Froot {
		_, qdir = dirProps(Fcwd)
	}
	qpaths, result, errlist := storeTransport(Fcwd, qdir)
	errs := make([]error, 0)
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}

	if len(errs) == 0 {
		qutil.EditList(Flist, Ftransported, qpaths)
		errs = nil
	}

	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}

func preSourceCo(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_co/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, true, false, "")

	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_co/2:\n", err)
		}
	}
}
