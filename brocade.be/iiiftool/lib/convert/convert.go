package convert

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	parallel "brocade.be/base/parallel"
)

var formatsAllowed = map[string]bool{".jpg": true, ".jpeg": true, ".tif": true}

// Convert files for IIIF in parallel using `gm` (GraphicsMagick)
// gm convert -flatten -quality 70 -define jp2:prg=rlcp -define jp2:numrlvls=7
// -define jp2:tilewidth=256 -define jp2:tileheight=256 s.tif o.jp2
func ConvertImageToJP2K(files []string, quality int, tile int) []error {

	fn := func(n int) (interface{}, error) {
		oldFile := files[n]
		ext := filepath.Ext(oldFile)
		_, found := formatsAllowed[ext]
		if !found {
			return nil, fmt.Errorf("file is not a valid image format: %v", oldFile)
		}

		newFile := filepath.Base(oldFile)
		newFile = strings.TrimSuffix(newFile, ext) + ".jp2"
		squality := strconv.Itoa(quality)
		stile := strconv.Itoa(tile)

		args := []string{"convert", "-flatten", "-quality", squality}
		args = append(args, "-define", "jp2:prg=rlcp", "-define", "jp2:numrlvls=7")
		args = append(args, "-define", "jp2:tilewidth="+stile, "-define", "jp2:tileheight="+stile)
		args = append(args, oldFile, newFile)

		cmd := exec.Command("gm", args...)
		_, err := cmd.Output()
		return newFile, err
	}

	_, errorlist := parallel.NMap(len(files), -1, fn)
	return errorlist
}
