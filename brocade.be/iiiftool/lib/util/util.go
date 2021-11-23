package util

import "strconv"

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
	// Specify input_file as - for standard input, output_file as - for standard output.
	// https://www.math.arizona.edu/~swig/documentation/ImgCvt/ImageMagick/www/convert.html
	args = append(args, "-", "-")
	return args
}
