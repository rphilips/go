package cmd

import (
	"log"

	"brocade.be/base/fs"
	"brocade.be/iiiftool/lib/convert"
	"github.com/spf13/cobra"
)

var fileConvertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert files for IIIF",
	Long: `Convert files for IIIF.
	Multiple files are handled in parallel.
`,
	Args:    cobra.MinimumNArgs(1),
	Example: `iiiftool file convert 1.jpg 2.jpg`,
	RunE:    fileConvert,
}

func init() {
	fileCmd.AddCommand(fileConvertCmd)
}

func fileConvert(cmd *cobra.Command, args []string) error {
	files := args
	for _, file := range files {
		if !fs.IsFile(file) {
			log.Fatalf("iiiftool ERROR: file is not valid: %v", file)
		}
	}
	err := convert.Run(files)
	if err != nil {
		log.Fatalf("iiiftool ERROR: error converting: %v", err)
	}

	return nil
}
