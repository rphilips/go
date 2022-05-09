package identifier

import (
	"io"
	"strings"
	"testing"

	"brocade.be/iiiftool/lib/util"
)

func TestCalculateIdentifier(t *testing.T) {
	files := []io.Reader{strings.NewReader("abc"), strings.NewReader("def")}
	manifest := `{"testing": "one, two"}`
	iiifsys := "foobar"
	efilesHash := "a4966be9021438019ed13e3c8e22551222bb1127"
	eidentifier := "6672b83be30e9c4956d6f5c82d896fe2937b7a4d"
	expected := efilesHash + "/" + eidentifier
	filesHash, identifier, _ := CalculateIdentifier(files, manifest, iiifsys)
	result := filesHash + "/" + identifier
	util.Check(result, expected, t)
}
