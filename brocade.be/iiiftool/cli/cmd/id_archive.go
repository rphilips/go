package cmd

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"brocade.be/base/docman"
	"brocade.be/base/parallel"
	"brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/sqlite"
	"brocade.be/iiiftool/lib/util"

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

func init() {
	idCmd.AddCommand(idArchiveCmd)
	idArchiveCmd.PersistentFlags().StringVar(&Furlty, "urlty", "", "URL type")
	idArchiveCmd.PersistentFlags().StringVar(&Fimgty, "imgty", "", "Image type")
	idArchiveCmd.PersistentFlags().StringVar(&Faccess, "access", "", "Access type")
	idArchiveCmd.PersistentFlags().StringVar(&Fmime, "mime", "", "Mime type")
	idArchiveCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "quality parameter")
	idArchiveCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "tile parameter")
}

func idArchive(cmd *cobra.Command, args []string) error {
	// verify input
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

	// harvest IIIF metadata from MUMPS
	result, err := iiif.Meta(id, loiType, Furlty, Fimgty, Faccess, Fmime)
	if err != nil {
		log.Fatalf("iiiftool ERROR: %s", err)
	}

	// get file contents from docman ids
	originalStream := make([]io.Reader, len(result.Images))
	originalfNames := make([]string, len(result.Images))

	empty := true
	for i, id := range result.Images {
		docid := docman.DocmanID(id)
		reader, err := docid.Reader()
		if err != nil {
			log.Fatalf("iiiftool ERROR: docman error:\n%s", err)
		}
		empty = false
		originalStream[i] = reader
		originalfNames[i] = filepath.Base(id)
	}
	if empty {
		log.Fatalf("iiiftool ERROR: no docman images found")
	}

	// convert file contents from TIFF/JPG to JP2K
	convertedStream := make([]io.Reader, len(originalStream))
	convertedfNames := make([]string, len(originalStream))

	fn := func(n int) (interface{}, error) {
		old := originalStream[n]
		args := util.GmConvertArgs(Fquality, Ftile)
		cmd := exec.Command("gm", args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return nil, err
		}
		out, err := cmd.StdoutPipe()
		if err != nil {
			return nil, err
		}
		_, err = cmd.StderrPipe()
		if err != nil {
			return nil, err
		}
		cmd.Start()

		go func() {
			defer stdin.Close()
			io.Copy(stdin, old)
		}()
		convertedStream[n] = out
		ext := filepath.Ext(originalfNames[n])
		convertedfNames[n] = strings.TrimSuffix(originalfNames[n], ext) + ".jp2"
		return out, nil
	}

	parallel.NMap(len(originalStream), -1, fn)

	// store file contents in SQLite archive
	filestream := make(map[string]io.Reader, len(convertedStream))
	for i, file := range convertedfNames {
		filestream[file] = convertedStream[i]
	}

	sqlitefile := id.Location(result.Digest)

	err = sqlite.Store(id, sqlitefile, filestream, Fcwd)
	if err != nil {
		log.Fatalf("iiiftool ERROR: store error:\n%s", err)
	}

	return nil
}
