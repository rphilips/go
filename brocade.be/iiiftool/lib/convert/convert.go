package convert

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"brocade.be/base/parallel"
	"brocade.be/iiiftool/lib/util"
)

var formatsAllowed = map[string]bool{".jpg": true, ".jpeg": true, ".tif": true}

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

	_, errors := parallel.NMap(len(files), -1, fn)
	return errors
}

// Convert streams for IIIF in parallel using `gm` (GraphicsMagick)
func ConvertStreamToJP2K(originalStream []io.Reader, quality int, tile int) ([]io.Reader, []error) {

	convertedStream := make([]io.Reader, len(originalStream))

	fn := func(n int) (interface{}, error) {
		old := originalStream[n]
		args := util.GmConvertArgs(quality, tile)
		// "Specify input_file as - for standard input, output_file as - for standard output",
		// so says http://www.graphicsmagick.org/GraphicsMagick.html#files,
		// but it needs to be "- jp2:-"!
		args = append(args, "-", "jp2:-")
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
		return out, nil
	}

	_, errors := parallel.NMap(len(originalStream), -1, fn)
	return convertedStream, errors
}
