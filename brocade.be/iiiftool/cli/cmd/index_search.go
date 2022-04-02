package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var indexSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search IIIF index",
	Long: `Search the IIIF index which stores the translation table
	for IIIF identifiers, IIIF digests and archive locations`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool index search`,
	RunE:    indexSearch,
}

func init() {
	indexCmd.AddCommand(indexSearchCmd)
}

func indexSearch(cmd *cobra.Command, args []string) error {

	search := args[0]
	if search == "" {
		log.Fatalf("iiiftool ERROR: argument is missing")
	}

	result, err := index.Search(search)

	if err != nil {
		log.Fatalf("iiiftool ERROR: error searching index:\n%s", err)
	}

	for _, res := range result {
		fmt.Println(res)
	}

	return nil
}
