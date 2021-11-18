package cmd

import (
	"fmt"
	"log"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idEncodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode a IIIF identifier",
	Long: `Encode a IIIF identifier using Base 32 encoding with URL and filename safe alphabet.
	Specification in RFC3548 (https://rfc-editor.org/rfc/rfc4648.html).
	`,

	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id encode dg:ua:1`,
	RunE:    idEncode,
}

func init() {
	idCmd.AddCommand(idEncodeCmd)
}

func idEncode(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])

	if id.String() == "" {
		log.Fatalf("argument is empty")
	}

	fmt.Println(id.Encode())
	return nil
}
