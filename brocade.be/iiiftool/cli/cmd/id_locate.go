package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var idLocateCmd = &cobra.Command{
	Use:     "locate",
	Short:   "Locate a IIIF id",
	Long:    `Locate a IIIF id in the index database.`,
	Args:    cobra.MinimumNArgs(1),
	Example: `iiiftool id locate dg:ua:9`,
	RunE:    idLocate,
}

func init() {
	idCmd.AddCommand(idLocateCmd)
}

func idLocate(cmd *cobra.Command, args []string) error {
	id := args[0]
	if id == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	search, err := index.Search(id)
	if err != nil {
		log.Fatalf("iiiftool ERROR: error searching index:\n%s", err)
	}

	// one id can be associated with several digests/locations
	for _, res := range search {
		location := res[3]
		fmt.Println(location)
	}
	return nil
}
