package cmd

import (
	"fmt"
	"log"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idDecodeCmd = &cobra.Command{
	Use:   "decode",
	Short: "Decode a IIIF identifier",
	Long: `Decode a IIIF identifier using Base 32 encoding with URL and filename safe alphabet.
	Specification in RFC3548 (https://rfc-editor.org/rfc/rfc4648.html).
	`,

	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id decode dg:ua:1`,
	RunE:    idDecode,
}

func init() {
	idCmd.AddCommand(idDecodeCmd)
}

func idDecode(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])

	if id.String() == "" {
		log.Fatalf("argument is empty")
	}

	fmt.Println(id.Decode())
	return nil
}
