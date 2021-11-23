package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"brocade.be/base/docman"
	"brocade.be/base/mumps"
	"brocade.be/base/parallel"
	identifier "brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/sqlite"

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
	fileCmd.PersistentFlags().IntVar(&Fquality, "quality", 70, "quality parameter")
	fileCmd.PersistentFlags().IntVar(&Ftile, "tile", 256, "tile parameter")
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

	convertedStream := make([]io.Reader, len(originalStream))
	convertedfNames := make([]string, len(originalStream))

	fn := func(n int) (interface{}, error) {
		old := originalStream[n]
		squality := strconv.Itoa(Fquality)
		stile := strconv.Itoa(Ftile)
		args := []string{"convert", "-flatten", "-quality", squality}
		args = append(args, "-define", "jp2:prg=rlcp", "-define", "jp2:numrlvls=7")
		args = append(args, "-define", "jp2:tilewidth="+stile, "-define", "jp2:tileheight="+stile)
		// Specify input_file as - for standard input, output_file as - for standard output.
		// https://www.math.arizona.edu/~swig/documentation/ImgCvt/ImageMagick/www/convert.html
		args = append(args, "-", "-")

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

	files := make(map[string]io.Reader, len(convertedStream))

	for i, file := range convertedfNames {
		files[file] = convertedStream[i]
	}

	err = sqlite.Store(id, files, Fcwd)
	if err != nil {
		log.Fatalf("iiiftool ERROR: store error:\n%s", err)
	}

	return nil
}
