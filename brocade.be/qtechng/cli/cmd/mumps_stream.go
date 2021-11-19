package cmd

import (
	"errors"
	"io"
	"os"
	"strings"

	qmumps "brocade.be/base/mumps"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var mumpsStreamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Starts a M stream",
	Long: `This command starts a M stream

The arguments are of the form key=value
The '--action' flag is mandatory and is an M expression`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng mumps stream loi=dg:ua:201 --action="d %Action^iiisori(.RApayload)"`,
	RunE:    mumpsStream,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BP",
		"fill-version":   "yes",
	},
}

var Faction = ""

func init() {
	mumpsStreamCmd.PersistentFlags().StringVar(&Faction, "action", "", "tag^routine indicating the M action")
	mumpsCmd.AddCommand(mumpsStreamCmd)
}

func mumpsStream(cmd *cobra.Command, args []string) error {
	if Faction == "" {
		Fmsg = qreport.Report(nil, errors.New("action flag is empty"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	payload := make(map[string]string)
	for _, arg := range args {
		key := ""
		value := ""
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key = parts[0]
			value = parts[1]
		} else {
			key = arg
		}
		if key == "" {
			Fmsg = qreport.Report(nil, errors.New("empty key"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		payload[key] = value

	}
	oreader, _, err := qmumps.Reader(Faction, payload)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	io.Copy(os.Stdout, oreader)

	return nil
}
