package cmd

import (
	"log"

	"brocade.be/iiiftool/lib/archive"
	"github.com/spf13/cobra"
)

var idArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Create archive for a IIIF identifier",
	Long: `Given a IIIF identifier, convert the appropriate image files
and store them in an SQLite archive.`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id archive c:stcv:12915850 --iiifsys=stcv`,
	RunE:    idArchive,
}

var Furlty string
var Fimgty string
var Faccess string
var Fmime string
var Fiiifsys string
var Findex bool
var Fmetaonly bool

func init() {
	idCmd.AddCommand(idArchiveCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Fiiifsys, "iiifsys", "test", "IIIF system")
	idArchiveCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "Quality parameter")
	idArchiveCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "Tile parameter")
	idArchiveCmd.PersistentFlags().BoolVar(&Findex, "index", true, "Rebuild IIIF index")
	idArchiveCmd.PersistentFlags().BoolVar(&Fverbose, "verbose", false, "Display information")
	idArchiveCmd.PersistentFlags().BoolVar(&Fmetaonly, "metaonly", false,
		`If images are present, only the meta information (including manifest) is replaced.
	If there are no images present, the usual archiving routine is performed.`)
}

func idArchive(cmd *cobra.Command, args []string) error {
	// Verify input
	id := args[0]
	if id == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	err := archive.Run(
		id,
		Fiiifsys,
		Fmetaonly,
		Findex,
		Fcwd,
		Fquality,
		Ftile,
		Fverbose)

	if err != nil {
		log.Fatalf("iiiftool ERROR: %s", err)
	}

	return nil
}
