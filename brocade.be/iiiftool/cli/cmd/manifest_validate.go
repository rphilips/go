package cmd

import (
	"fmt"
	"log"

	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var manifestValidateCmd = &cobra.Command{
	Use:     "validate",
	Short:   "Validate manifest for a IIIF identifier",
	Long:    `Validate a IIIF manifest URL given a certain version`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool manifest validate https://dev.anet.be/iiif/e0f4d5d32a3dd5a341ec84a2ae8e9c69e2666fca/manifest --version=2.1`,
	RunE:    manifestValidate,
}

var Fversion string

func init() {
	manifestCmd.AddCommand(manifestValidateCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Fversion, "version", "3.0", "IIIF Presentation API version")
}

func manifestValidate(cmd *cobra.Command, args []string) error {
	url := args[0]
	if url == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	result, err := iiif.Validate(url, Fversion)
	if err != nil {
		log.Fatalf("iiiftool ERROR: error validating: %v", err)
	}

	fmt.Println(report.Report(result, nil, []string{""}, false, false, "", false, "", ""))

	return nil
}
