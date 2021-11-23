package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	identifier "brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/sqlite"

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
	If --cwd is used, the archive is created in the specified working directory.
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

	files := make(map[string]io.Reader, len(args[1:]))

	for _, file := range args[1:] {
		name := filepath.Base(file)
		reader, err := os.Open(file)
		if err != nil {
			log.Fatalf("iiiftool ERROR: file is not valid: %v", file)
		}
		files[name] = reader
	}

	err := sqlite.Store(id, files, Fcwd)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot store:\n%s", err)
	}

	return nil
}
