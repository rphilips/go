package cmd

import (
	"log"
	"strconv"

	fs "brocade.be/base/fs"
	"brocade.be/iiiftool/lib/sqlite"
	"github.com/spf13/cobra"
)

var digestHarvestCmd = &cobra.Command{
	Use:   "harvest",
	Short: "Harvest from a IIIF digest",
	Long: `Given a IIIF digest and a file,
	harvest that file from the appropriate SQLite archive.`,
	Args:    cobra.MinimumNArgs(2),
	Example: `iiiftool digest harvest a42f98d253ea3dd019de07870862cbdc62d6077c 00000001.jp2`,
	RunE:    digestHarvest,
}

func init() {
	digestCmd.AddCommand(digestHarvestCmd)
}

func digestHarvest(cmd *cobra.Command, args []string) error {
	digest := args[0]
	if digest == "" {
		log.Fatalf("iiiftool ERROR: digest is empty")
	}
	file := args[1]
	if file == "" {
		log.Fatalf("iiiftool ERROR: file is empty")
	}

	harvestcode := digest + file
	var sqlar sqlite.Sqlar
	err := sqlite.Harvest(harvestcode, &sqlar)
	if err != nil {
		log.Fatalf("iiiftool ERROR: file cannot be harvested\n%s", err)
	}

	err = fs.Store(file, sqlar.Reader, strconv.Itoa(sqlar.Mode))
	if err != nil {
		log.Fatalf("iiiftool ERROR: file cannot be created\n%s", err)
	}

	return nil
}
