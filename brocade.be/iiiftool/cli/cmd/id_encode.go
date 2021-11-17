package cmd

import (
	"errors"
	"fmt"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idEncodeCmd = &cobra.Command{
	Use:   "encode",
	Short: "Encode a IIIF identifier",
	Long: `Encode a IIIF identifier using Base 64 encoding with URL and filename safe alphabet.
	Specification in RFC4648 (https://datatracker.ietf.org/doc/html/rfc4648#page-7)
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
		return errors.New("argument is empty")
	}

	fmt.Println(id.Encode())
	return nil
}