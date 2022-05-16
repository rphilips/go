package cmd

import (
	"log"

	"brocade.be/iiiftool/lib/archive"
	"brocade.be/iiiftool/lib/index"
	"github.com/spf13/cobra"
)

var manifestUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update manifest for a IIIF identifier",
	Long:  `Given a IIIF identifier and IIIF system update the manifest`,
	Args:  cobra.RangeArgs(0, 2),
	Example: `iiiftool manifest update c:stcv:12915850 stcv
	iiiftool manifest update --all`,
	RunE: manifestUpdate,
}

var Fall bool

func init() {
	manifestCmd.AddCommand(manifestUpdateCmd)
	manifestUpdateCmd.PersistentFlags().BoolVar(&Fall, "all", false, "Update the complete IIIF database")
}

func manifestUpdate(cmd *cobra.Command, args []string) error {

	if Fall {
		indexRows, err := index.LookupAll()
		if err != nil {
			log.Fatalf("iiiftool ERROR: cannot lookup identifiers: %v", err)
		}

		digests := make(map[string]bool, len(indexRows))

		for _, indexRow := range indexRows {
			id := indexRow[1]
			digest := indexRow[2]
			iiifsys := indexRow[3]

			if digests[digest] {
				continue
			}

			digests[digest] = true
			err := archive.Run(id, iiifsys, true, true, "", 0, 0, false)
			if err != nil {
				log.Fatalf("iiiftool ERROR: cannot update archive: %v", err)
			}

		}

		return nil
	}

	if len(args) != 0 {
		if len(args) < 2 {
			log.Fatalf("iiiftool ERROR: too few arguments")
		}
		id := args[0]
		iiifsys := args[1]

		// image parameters can be 0 because there is never image conversion
		err := archive.Run(id, iiifsys, true, true, "", 0, 0, false)
		if err != nil {
			log.Fatalf("iiiftool ERROR: cannot update archive: %v", err)
		}

		return nil

	}

	return nil
}
