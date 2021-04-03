package cmd

import (
	"log"
	"path"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var sourceCoCmd = &cobra.Command{
	Use:     "co",
	Short:   "Checks out QtechNG files",
	Long:    `Command to retrieve files from the QtechNG repository`,
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

// Flist identifier of list of the results, if in auto mode
var Flist string

func init() {
	sourceCoCmd.Flags().BoolVar(&Fclear, "clear", false, "Clears visited directories, if in auto mode")
	sourceCoCmd.Flags().StringVar(&Flist, "list", "", "List with qpaths, if in auto mode")
	sourceCmd.AddCommand(sourceCoCmd)
}

func sourceCo(cmd *cobra.Command, args []string) error {
	result, errlist := storeTransport()
	if len(errlist) == 0 {

		supportdir := qregistry.Registry["qtechng-support-dir"]
		if Flist != "" && supportdir != "" {
			lst := make([]string, len(result))
			for i, st := range result {
				lst[i] = st.QPath
			}
			if len(lst) != 0 {
				listname := path.Join(supportdir, "data", Flist+".lst")
				qfs.Mkdir(path.Dir(listname), "process")
				qfs.Store(listname, strings.Join(lst, "\n"), "process")
			}
		}

		Fmsg = qerror.ShowResult(result, Fjq, nil, Fyaml)
		return nil
	}
	Fmsg = qerror.ShowResult(result, Fjq, qerror.ErrorSlice(errlist), Fyaml)
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
		addData(Fpayload, Fcargo, true, "")

	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_co/2:\n", err)
		}
	}
}
