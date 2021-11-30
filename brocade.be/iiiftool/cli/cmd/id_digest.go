package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/index"

	"github.com/spf13/cobra"
)

var idDigestCmd = &cobra.Command{
	Use:     "digest",
	Short:   "Look up digest for a IIIF identifier",
	Long:    `Given a IIIF identifier, find its digest in the index database`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id digest dg:ua:9`,
	RunE:    idDigest,
}

func init() {
	idCmd.AddCommand(idDigestCmd)
}

func idDigest(cmd *cobra.Command, args []string) error {
	id := args[0]
	if id == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	digest, err := index.LookupId(id)
	if err != nil {
		log.Fatalf("iiiftool ERROR: error looking up digest:%s", err)
	}
	fmt.Println(digest)

	return nil
}
