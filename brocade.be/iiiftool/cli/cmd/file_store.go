package cmd

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/sqlite"
	"github.com/spf13/cobra"
)

var fileStoreCmd = &cobra.Command{
	Use:   "store",
	Short: "Store files for IIIF",
	Long: `Store files for IIIF in an SQLite archive.
	The first argument is the SQLite archive,
	the other arguments are the files to store.
	If the files append already exist in the archive,
	they are appended.
`,
	Args:    cobra.MinimumNArgs(2),
	Example: `iiiftool file store mydb.sqlite 1.jp2 2.jp2 dg_ua_1.json`,
	RunE:    fileStore,
}

func init() {
	fileCmd.AddCommand(fileStoreCmd)
}

func fileStore(cmd *cobra.Command, args []string) error {
	sqlitefile := args[0]

	if sqlitefile == "" {
		log.Fatalf("iiiftool ERROR: SQLite archive is missing")
	}

	var dummyMeta iiif.MResponse

	files := make([]io.Reader, len(args[1:]))

	for i, file := range args[1:] {
		reader, err := os.Open(file)
		if err != nil {
			log.Fatalf("iiiftool ERROR: file is not valid: %v", file)
		}
		files[i] = reader

		name := filepath.Base(file)
		data := map[string]string{"name": name}
		dummyMeta.Images = append(dummyMeta.Images, data)

	}

	err := sqlite.Store(sqlitefile, files, Fcwd, dummyMeta)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot store:\n%s", err)
	}

	return nil
}
