package cmd

import (
	"log"

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
	If an SQLite archive for the identifier already exists,
	the files are appended.
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
		log.Fatalf("iiiftool ERROR: identifier is missing")
	}

	err := sqlite.Store(id, args[1:])
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot store:\n%s", err)
	}

	return nil
}
