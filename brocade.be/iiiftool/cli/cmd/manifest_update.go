package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"brocade.be/base/mumps"
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
iiiftool manifest update --all
iiiftool manifest update --modified`,
	RunE: manifestUpdate,
}

var Fall bool
var Fmodified bool
var Fdry bool

func init() {
	manifestCmd.AddCommand(manifestUpdateCmd)
	manifestUpdateCmd.PersistentFlags().BoolVar(&Fall, "all", false, "Update the complete IIIF database")
	manifestUpdateCmd.PersistentFlags().BoolVar(&Fverbose, "verbose", false, "Show verbose output")
	manifestUpdateCmd.PersistentFlags().BoolVar(&Fdry, "dry", false, "Dry run: show output, but do not perform update")
	manifestUpdateCmd.PersistentFlags().BoolVar(&Fmodified, "modified", false, `Update all IIIF digests connected to LOIs
that have been modified since the creation of their manifest`)
}

func manifestUpdate(cmd *cobra.Command, args []string) error {

	if Fdry {
		Fverbose = true
	}

	if Fmodified {
		payload := make(map[string]string)
		oreader, _, err := mumps.Reader("d %Modif^gbiiif(.RApayload)", payload)
		if err != nil {
			log.Fatalf("mumps reader error:\n%s", err)
		}

		out, err := ioutil.ReadAll(oreader)
		if err != nil {
			log.Fatalf("cannot read MUMPS response:\n%s", err)
		}

		result := make(map[string]map[string]string)

		err = json.Unmarshal(out, &result)
		if err != nil {
			log.Fatalf("json error:\n%s", err)
		}

		for digest, data := range result {
			id := data["loi"]
			iiifsys := data["iiifsys"]

			err := archive.Update(digest, id, iiifsys, Fverbose, Fdry)
			if err != nil {
				log.Fatalf("iiiftool ERROR: cannot update archive: %v", err)
			}
		}

		return nil
	}

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

			err := archive.Update(digest, id, iiifsys, Fverbose, Fdry)
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
		err := archive.Update("", id, iiifsys, Fverbose, Fdry)
		if err != nil {
			log.Fatalf("iiiftool ERROR: cannot update archive: %v", err)
		}

		return nil

	}

	return nil
}
