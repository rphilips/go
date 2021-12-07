package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var idLocateCmd = &cobra.Command{
	Use:   "locate",
	Short: "Locate a IIIF id",
	Long: `Given a IIIF id formulate an appropriate SQLite filepath
-- regardless of whether this location actually exists or not.`,
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

	result := make(map[string]bool)

	for _, res := range search {
		location := res[3]
		result[location] = true
	}

	for location := range result {
		fmt.Println(location)
	}
	return nil
}
