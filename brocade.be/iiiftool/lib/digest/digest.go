package digest

import (
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	"brocade.be/iiiftool/lib/util"
)

func CalculateDigest(
	files []io.Reader,
	manifest string,
	iiifsys string) (string, string, error) {

	// Van elk van de bestanden uit sqlite archief wordt een hexadecimale lowercase SHA-1 berekend (cryptografische robuustheid is niet vereist).
	// Deze lijst van hashes wordt lexicografisch gesorteerd
	// De gesorteerde hashes worden geconcateneerd
	// Van deze string wordt opnieuw een hexadecimale lowercase SHA-1 berekend
	// deze waarde H1 wordt opgeslagen in de SQLite
	// er wordt een hexadecimale lowercase SHA-1 H2 berekend van het manifest (as-is)
	// naam van het IIIF-systeem, H1 en H2 worden in deze volgorde geconcateneerd
	// de digest is de hexadecimale lowercase SHA-1 van deze nieuwe string

	// files
	fileHashes := make([]string, len(files))
	for _, file := range files {
		data, err := ioutil.ReadAll(file)
		if err != nil {
			return "", "", fmt.Errorf("cannot read stream: %v", err)
		}
		fileHashes = append(fileHashes, util.GetSHA1(data))
	}
	sort.Slice(fileHashes, func(i, j int) bool { return fileHashes[i] < fileHashes[j] })
	filesHash := strings.Join(fileHashes, "")
	filesHash = util.GetSHA1([]byte(filesHash))

	// manifest
	manifestHash := util.GetSHA1([]byte(manifest))

	// iiifsys
	result := iiifsys + filesHash + manifestHash

	digest := util.GetSHA1([]byte(result))

	return filesHash, digest, nil
}
