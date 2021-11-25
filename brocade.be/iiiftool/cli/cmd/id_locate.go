package cmd

import (
	"fmt"
	"log"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idLocateCmd = &cobra.Command{
	Use:   "locate",
	Short: "Locate a IIIF identifier",
	Long: `Given a IIIF identifier formulate an appropriate SQLite filepath.
	You can choose a digest to use for generating the path,
	or have the system generate it from scratch`,
	Args: cobra.MinimumNArgs(1),
	Example: `iiiftool id locate dg:ua:1
	iiiftool id locate dg:ua:1 a42f98d253ea3dd019de07870862cbdc62d6077c`,
	RunE: idLocate,
}

func init() {
	idCmd.AddCommand(idLocateCmd)
}

func idLocate(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])
	if id.String() == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	if len(args) < 2 {
		args = append(args, "")
	}
	digest := args[1]

	fmt.Println(id.Location(digest))
	return nil
}
