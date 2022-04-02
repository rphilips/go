package cmd

import (
	"log"

	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var indexRebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild IIIF index",
	Long: `Rebuild the IIIF index which stores the translation table
	for IIIF identifiers and IIIF digests`,
	Args:    cobra.NoArgs,
	Example: `iiiftool index rebuild`,
	RunE:    indexRebuild,
}

func init() {
	indexCmd.AddCommand(indexRebuildCmd)
	indexRebuildCmd.PersistentFlags().BoolVar(&Fverbose, "verbose", false, "Display information")
}

func indexRebuild(cmd *cobra.Command, args []string) error {

	err := index.Rebuild(Fverbose)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot rebuild index:\n%s", err)
	}

	return nil
}
