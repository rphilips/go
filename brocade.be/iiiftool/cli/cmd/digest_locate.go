package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/iiif"
	"github.com/spf13/cobra"
)

var digestLocateCmd = &cobra.Command{
	Use:   "locate",
	Short: "Locate a IIIF digestentifier",
	Long: `Given a IIIF digestentifier formulate an appropriate SQLite filepath.
	You can choose a digest to use for generating the path,
	or have the system generate it from scratch`,
	Args:    cobra.MinimumNArgs(1),
	Example: `iiiftool digest locate a42f98d253ea3dd019de07870862cbdc62d6077c`,
	RunE:    digestLocate,
}

func init() {
	digestCmd.AddCommand(digestLocateCmd)
}

func digestLocate(cmd *cobra.Command, args []string) error {
	digest := args[0]
	if digest == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	fmt.Println(iiif.Digest2Location(digest))
	return nil
}
