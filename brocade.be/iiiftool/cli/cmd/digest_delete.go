package cmd

import (
	"log"

	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var digestDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a IIIF digest",
	Long: `Delete a IIIF digest together with all its (meta)data,
	i.e. SQLite archive and index entries`,
	Args:    cobra.MinimumNArgs(1),
	Example: `iiiftool digest delete a42f98d253ea3dd019de07870862cbdc62d6077c`,
	RunE:    digestdelete,
}

func init() {
	digestCmd.AddCommand(digestDeleteCmd)
}

func digestdelete(cmd *cobra.Command, args []string) error {
	digest := args[0]
	if digest == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	err := iiif.DigestDelete(digest)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot delete digest\n%s", err)
	}

	err = index.RemoveDigest(digest)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot delete index entry\n%s", err)
	}

	return nil
}
