package cmd

import (
	"fmt"
	"log"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idLocateCmd = &cobra.Command{
	Use:     "locate",
	Short:   "Locate a IIIF identifier",
	Long:    "Given a IIIF identifier locate the appropriate SQLite filepath",
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id locate dg:ua:1`,
	RunE:    idLocate,
}

var Freverse bool

func init() {
	idCmd.AddCommand(idLocateCmd)
	idLocateCmd.PersistentFlags().BoolVar(&Freverse, "reverse", false, "Reconstruct the IIIF identifier from the filepath")
}

func idLocate(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])

	if id.String() == "" {
		log.Fatalf("argument is empty")
	}
	if !Freverse {
		fmt.Println(id.Location())
	} else {
		fmt.Println(identifier.ReverseLocation(args[0]))
	}
	return nil
}
