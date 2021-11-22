package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	docman "brocade.be/base/docman"
	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idArchiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Create archive for a IIIF identifier",
	Long: `Given a IIIF identifier, put the appropriate image files in an SQLite archive.
Various additional parameters are in use and sometimes required:
--urlty:	url type (required for c-loi/o-loi)
--imgty:	image type (required for tg-loi)
--access:	access type (space separated)
--mime:		mime type (space separated)`,
	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id archive dg:ua:1`,
	RunE:    idArchive,
}

var Furlty = ""
var Fimgty = ""
var Faccess = ""
var Fmime = ""
var qArgs = []string{"mumps", "stream"}
var paths []string

type mPayload struct {
	Identifier string
	Images     []string
}

func init() {
	idCmd.AddCommand(idArchiveCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Furlty, "urlty", "", "URL type")
	idArchiveCmd.PersistentFlags().StringVar(&Fimgty, "imgty", "", "Image type")
	idArchiveCmd.PersistentFlags().StringVar(&Faccess, "access", "", "Access type")
	idArchiveCmd.PersistentFlags().StringVar(&Fmime, "mime", "", "Mime type")
}

func idArchive(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])
	if id.String() == "" {
		log.Fatalf("iiiftool ERROR: argument is empty")
	}

	loiType := strings.Split(id.String(), ":")[0]
	switch loiType {
	case "c", "o":
		if Furlty == "" {
			log.Fatalf("iiiftool ERROR: c-loi requires --urlty flag")
		}
	case "tg":
		if Fimgty == "" {
			log.Fatalf("iiiftool ERROR: c-loi requires --urlty flag")
		}
	}

	qArgs = append(qArgs, "loi="+id.String())
	switch loiType {
	case "c", "o":
		qArgs = append(qArgs, "urlty="+Furlty)
	case "tg":
		qArgs = append(qArgs, "imgty="+Fimgty)
	}
	if Faccess != "" {
		qArgs = append(qArgs, "access="+Faccess)
	}
	if Fmime != "" {
		qArgs = append(qArgs, "mime="+Fmime)
	}
	qArgs = append(qArgs, "--action=d %Action^iiisori(.RApayload)")

	// qcmd := exec.Command("qtechng", qArgs...)
	// out, err := qcmd.Output()
	// if err != nil {
	// 	log.Fatalf("iiiftool ERROR: mumps error:\n%s", err)
	// }
	fmt.Println(qArgs)
	out := []byte(`{"identifier": "dg:ua:201", "images":["/uact/255909/1.tif","/uact/b6199c/2.tif","/uact/1fed6c/3.tif"]}`)
	var result mPayload
	json.Unmarshal([]byte(out), &result)

	for _, id := range result.Images {
		path := docman.DocmanID(id)
		paths = append(paths, path.Location())
	}
	fmt.Println(paths)
	// iterate over map: get filepath from docman ids, put filepaths in slice
	// store files in archive

	return nil
}
