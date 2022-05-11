package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"brocade.be/base/fs"
	"brocade.be/iiiftool/lib/iiif"

	"github.com/spf13/cobra"
)

var manifestGenerateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "Generate manifest for a IIIF identifier",
	Long:    `Given a IIIF identifier and IIIF system generate the manifest`,
	Args:    cobra.ExactArgs(2),
	Example: `iiiftool manifest generate c:stcv:12915850 stcv`,
	RunE:    manifestGenerate,
}

func init() {
	manifestCmd.AddCommand(manifestGenerateCmd)
}

func manifestGenerate(cmd *cobra.Command, args []string) error {
	loi := args[0]
	if loi == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}
	iiifsys := args[1]
	if iiifsys == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	iiifMeta, err := iiif.Meta(loi, iiifsys)
	if err != nil {
		log.Fatalf("iiiftool ERROR: %s", err)
	}

	manifest, err := json.Marshal(iiifMeta.Manifest)
	if err != nil {
		return fmt.Errorf("json error:\n%s", err)
	}

	err = fs.Store("manifest.json", manifest, "nakedfile")
	if err != nil {
		return fmt.Errorf("json error:\n%s", err)
	}

	return nil
}
