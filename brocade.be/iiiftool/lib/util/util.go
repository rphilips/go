package util

import (
	"fmt"
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
