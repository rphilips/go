package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// Function that takes a string as argument
// and returns the reverse of string.
func StrReverse(str string) (result string) {
	for _, v := range str {
		result = string(v) + result
	}
	return result
}

// Function that prepares gm conversion command arguments
func GmConvertArgs(quality int, tile int) []string {
	squality := strconv.Itoa(quality)
	stile := strconv.Itoa(tile)
	args := []string{"convert", "-flatten", "-quality", squality}
	args = append(args, "-define", "jp2:prg=rlcp", "-define", "jp2:numrlvls=7")
	args = append(args, "-define", "jp2:tilewidth="+stile, "-define", "jp2:tileheight="+stile)
	return args
}

// Make string URL-safe
func URLSafe(data string) string {
	data = strings.ToLower(data)
	unsafeRegexp := regexp.MustCompile(`[^a-z0-9]`)
	data = unsafeRegexp.ReplaceAllString(data, "_")
	return data
}

// Compare result and expected for tests
func Check(result string, expected string, t *testing.T) {
	if result != expected {
		t.Errorf(fmt.Sprintf("\nResult: \n[%s]\nExpected: \n[%s]\n", result, expected))
	}
}

// Get SHA1 for a given string
func GetSHA1(data []byte) string {
	h := sha1.New()
	h.Write(data)
	sha1bytes := h.Sum(nil)
	hash := hex.EncodeToString(sha1bytes)
	hash = strings.ToLower(hash)
	return hash
}

// Get unique values from a slice of identifiers like this
// [dg:ua:100 dg:ua:100,iiifsys:uapr tg:uact:25]
func GetUniqueLOIs(data []string) map[string]bool {
	result := make(map[string]bool)
	for _, loi := range data {
		loi = strings.Split(loi, ",")[0]
		loi = strings.TrimRight(loi, ",")
		result[loi] = true
	}
	return result
}

// Create file with full nested path
func CreateFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	file, err := os.Create(p)
	return file, err
}
