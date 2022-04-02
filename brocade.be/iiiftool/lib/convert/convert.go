package convert

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"brocade.be/base/docman"
	"brocade.be/base/parallel"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/util"
)

var formatsAllowed = map[string]bool{".jpg": true, ".jpeg": true, ".tif": true}
var iiifMaxPar, _ = strconv.Atoi(registry.Registry["iiif-max-parallel"])

// Convert files for IIIF in parallel using `gm` (GraphicsMagick)
// gm convert -flatten -quality 70 -define jp2:prg=rlcp -define jp2:numrlvls=7 -define jp2:tilewidth=256 -define jp2:tileheight=256 s.tif o.jp2
func ConvertFileToJP2K(files []string, quality int, tile int, cwd string) []error {

	fn := func(n int) (interface{}, error) {
		oldFile := files[n]
		ext := filepath.Ext(oldFile)
		_, found := formatsAllowed[ext]
		if !found {
			return nil, fmt.Errorf("file is not a valid image format: %v", oldFile)
		}

		newFile := filepath.Base(oldFile)
		newFile = strings.TrimSuffix(newFile, ext) + ".jp2"
		args := util.GmConvertArgs(quality, tile)
		if cwd != "" {
			newFile = filepath.Join(cwd, newFile)
		}
		args = append(args, oldFile, newFile)

		cmd := exec.Command("gm", args...)
		_, err := cmd.Output()
		return newFile, err
	}

	_, errors := parallel.NMap(len(files), iiifMaxPar, fn)
	return errors
}

// Convert docman ids for IIIF in parallel using `gm` (GraphicsMagick)
func ConvertDocmanIdsToJP2K(docIds []docman.DocmanID, quality int, tile int, verbose bool) ([]io.Reader, []error) {

	convertedStream := make([]io.Reader, len(docIds))

	fn := func(n int) (interface{}, error) {
		if verbose {
			fmt.Println(docIds[n])
		}
		old, err := docIds[n].Reader()
		if err != nil {
			return nil, err
		}

		args := util.GmConvertArgs(quality, tile)
		// "Specify input_file as - for standard input, output_file as - for standard output",
		// so says http://www.graphicsmagick.org/GraphicsMagick.html#files,
		// but it needs to be "- jp2:-"!
		args = append(args, "-", "jp2:-")
		cmd := exec.Command("gm", args...)
		cmd.Stdin = old

		_, err = cmd.StderrPipe()
		if err != nil {
			return nil, err
		}

		blob, err := cmd.Output()
		old.Close()
		if err != nil {
			return nil, err
		}

		convertedStream[n] = bytes.NewReader(blob)

		return nil, nil
	}

	_, errors := parallel.NMap(len(docIds), iiifMaxPar, fn)
	return convertedStream, errors
}
