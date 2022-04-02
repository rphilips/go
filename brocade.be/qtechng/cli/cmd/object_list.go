package cmd

import (
	"encoding/json"
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var objectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List objects in the repository",
	Long: `This command lists Brocade objects (i4/l4/m4/r4/t4) in the repository
The objects can be specified:
    - as arguments
	- as '--objpattern-...' flags.

Do not forget the appropriate prefix!`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng object list l4_loi l4_title
qtechng object list --objpattern='m4_getCat*'
	`,
	RunE:   objectList,
	PreRun: preObjectList,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	objectCmd.AddCommand(objectListCmd)
	objectListCmd.PersistentFlags().StringArrayVar(&Fobjpattern, "objpattern", []string{}, "Posix glob pattern on object names")
}

func objectList(cmd *cobra.Command, args []string) error {

	result := listObjectTransport(Fcargo)
	v := make(map[string]interface{})
	json.Unmarshal(result, &v)
	Fmsg = qreport.Report(v, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}

func preObjectList(cmd *cobra.Command, args []string) {

	if !Ftransported {
		var err error
		Fcargo, err = fetchObjectData(args)
		if err != nil {
			log.Fatal("cmd/object_list/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addObjectData(Fpayload, Fcargo, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/object_list/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
