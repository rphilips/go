package cmd

import (
	"errors"

	identifier "brocade.be/iiiftool/lib/identifier"
	sqlite "brocade.be/iiiftool/lib/sqlite"

	"github.com/spf13/cobra"
)

var fileStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store files for IIIF",
	Long: `Store files for IIIF in an SQLite archive.
	The first argument is the IIIF identifier,
	the other arguments are the files to store.
`,
	Args:    cobra.MinimumNArgs(2),
	Example: `iiiftool file store dg:ua:1 1.jp2 2.jp2 dg_ua_1.json`,
	RunE:    fileStore,
}

func init() {
	fileCmd.AddCommand(fileStoreCmd)
}

func fileStore(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])

	if id.String() == "" {
		return errors.New("identifier is missing")
	}

	_ = sqlite.Store(id, args[1:])

	return nil
}
