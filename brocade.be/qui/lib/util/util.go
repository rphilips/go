package util

import (
	"regexp"
	"strings"
)

func ToHTML(data string) string {
	data = strings.ReplaceAll(data, "\n", "<br>")
	linkRegex := regexp.MustCompile(`http.*?:\/\/.*?<`)
	matches := linkRegex.FindAllString(data, -1)
	for _, match := range matches {
		ref := strings.Trim(match, "<")
		data = strings.ReplaceAll(data, ref, `<a href="`+ref+`">`+ref+`</a>`)
	}
	// trick to render indentation in HTML
	data = strings.ReplaceAll(data, "  ", "<span style=color:white>.</span>")

	return data
}
