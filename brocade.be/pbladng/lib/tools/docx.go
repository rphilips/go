package tools

import (
	"strings"

	docx "github.com/nguyenthenguyen/docx"
)

func Grep(fname string, needle string) (found bool, err error) {
	r, err := docx.ReadDocxFile(fname)
	if err != nil {
		return
	}
	body := r.Editable().GetContent()
	return strings.Contains(body, needle), nil
}
