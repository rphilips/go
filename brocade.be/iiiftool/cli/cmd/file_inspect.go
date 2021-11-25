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
	- meta (IIIF meta information)`,
	Args:    cobra.MinimumNArgs(2),
	Example: `iiiftool file inspect mydb.sqlite sqlar`,
	RunE:    fileInspect,
}

func init() {
	fileCmd.AddCommand(fileInspectCmd)
}

func fileInspect(cmd *cobra.Command, args []string) error {

	result, err := sqlite.Inspect(args[0], args[1])
	if err != nil {
		log.Fatalf("iiiftool ERROR: cannot inspect:\n%s", err)
	}
	fmt.Println(result)

	return nil
}
