package cmd

import (
	"log"

	fs "brocade.be/base/fs"
	convert "brocade.be/iiiftool/lib/convert"
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

var Fquality int
var Ftile int

func init() {
	fileCmd.AddCommand(fileConvertCmd)
	fileCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "quality parameter")
	fileCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "tile parameter")
}

func fileConvert(cmd *cobra.Command, args []string) error {
	files := args
	for _, file := range files {
		if !fs.IsFile(file) {
			log.Fatalf("iiiftool ERROR: file is not valid: %v", file)
		}
	}
	err := convert.ConvertImageToJP2K(files, Fquality, Ftile)
	for _, e := range err {
		if e != nil {
			log.Fatalf("iiiftool ERROR: error converting: %v", e)
		}
	}

	return nil
}
