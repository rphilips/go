package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/sqlite"
	"github.com/spf13/cobra"
)

var digestManifestCmd = &cobra.Command{
	Use:   "manifest",
	Short: "Show manifest for a IIIF digest",
	Long: `Given a IIIF digest, show the manifest
	that is contained in the appropriate SQLite archive.`,
	Args:    cobra.MinimumNArgs(1),
	Example: `iiiftool digest manifest a42f98d253ea3dd019de07870862cbdc62d6077c`,
	RunE:    digestManifest,
}

func init() {
	digestCmd.AddCommand(digestManifestCmd)
}

func digestManifest(cmd *cobra.Command, args []string) error {
	digest := args[0]
	if digest == "" {
		log.Fatalf("iiiftool ERROR: digest is empty")
	}

	manifest, err := sqlite.Manifest(digest)
	if err != nil {
		log.Fatalf("iiiftool ERROR: manifest cannot be retrieved\n%s", err)
	}

	fmt.Println(manifest)

	return nil
}
