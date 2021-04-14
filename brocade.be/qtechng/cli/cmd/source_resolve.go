package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	"github.com/spf13/cobra"
)

//Frilm r4/i4/l4/m4 substitutie
var Frilm string = ""

var sourceResolveCmd = &cobra.Command{
	Use:     "resolve",
	Short:   "Resolves a sourcefile",
	Long:    `Resolves the i4/r4/m4/l4 constructions in sources`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source resolves --pattern=/catalografie/application/bcawedit.m`,
	RunE:    sourceResolve,
	PreRun:  preSourceResolve,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceResolveCmd.PersistentFlags().StringVar(&Frilm, "rilm", "", "specify the substitutions")
	sourceCmd.AddCommand(sourceResolveCmd)
}

func sourceResolve(cmd *cobra.Command, args []string) error {
	Fmsg = string(Fcargo.Data)
	return nil
}

func preSourceResolve(cmd *cobra.Command, args []string) {
	if Frilm == "" {
		Frilm = "rilm"
	}
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_resolve/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, true, "r:"+Frilm)

	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_resolve/2:\n", err)
		}
	}
}
