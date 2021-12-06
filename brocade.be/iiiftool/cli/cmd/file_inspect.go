package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/sqlite"

	"github.com/spf13/cobra"
)

var fileInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect files for IIIF",
	Long: `Inspect a SQLite archive.
	The first argument is the SQLite archive, the second is the table to inspect:
	- admin (administrative info)
	- sqlar (archive)
	- files (file info)
	- meta (IIIF meta information)
	If no second argument is supplied, inspect shows the database schema`,
	Args: cobra.MinimumNArgs(1),
	Example: `iiiftool file inspect mydb.sqlite sqlar
iiiftool file inspect mydb.sqlite`,
	RunE: fileInspect,
}

func init() {
	fileCmd.AddCommand(fileInspectCmd)
}

func fileInspect(cmd *cobra.Command, args []string) error {

	sqliteFile := args[0]
	if sqliteFile == "" {
		log.Fatalf("iiiftool ERROR: argument 1 is missing")
	}
	table := ""
	if len(args) == 2 {
		table = args[1]
	}

	result, err := sqlite.Inspect(sqliteFile, table)
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot inspect:\n%s", err)
	}
	fmt.Println(result)

	return nil
}
