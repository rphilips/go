package cmd

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

	docman "brocade.be/base/docman"
	mumps "brocade.be/base/mumps"
	identifier "brocade.be/iiiftool/lib/identifier"
	sqlite "brocade.be/iiiftool/lib/sqlite"

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

type mResponse struct {
	Identifier string
	Images     []string
}

// puur resultaat in --cwd

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
			log.Fatalf("iiiftool ERROR: tg-loi requires --imgty flag")
		}
	}

	payload := make(map[string]string)
	payload["loi"] = id.String()
	switch loiType {
	case "c", "o":
		payload["urlty"] = Furlty
	case "tg":
		payload["imgty"] = Fimgty
	}
	if Faccess != "" {
		payload["access"] = Faccess
	}
	if Fmime != "" {
		payload["mime"] = Fmime
	}

	oreader, _, err := mumps.Reader("d %Action^iiisori(.RApayload)", payload)
	if err != nil {
		log.Fatalf("iiiftool ERROR: mumps error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		log.Fatalf("iiiftool ERROR: mumps error:\n%s", err)
	}
	var result mResponse
	json.Unmarshal(out, &result)
	paths := make([]string, len(result.Images))

	for i, id := range result.Images {
		path := docman.DocmanID(id)
		paths[i] = path.Location()
	}
	err = sqlite.Store(id, paths, Fcwd)
	if err != nil {
		log.Fatalf("iiiftool ERROR: store error:\n%s", err)
	}

	return nil
}
