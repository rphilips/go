package util

import (
	"fmt"
	"strconv"
	"strings"
)

// ApplyAlgo transform to fit in a string
func ApplyAlgo(data string, algo string) (result string) {
	if data == "" {
		return
	}
	if algo == "" {
		return data
	}
	switch algo {
	case "js":
		result = applyJS(data)
	case "py":
		result = applyPY(data)
	case "php":
		result = applyPHP(data)
	case "xml":
		result = applyXML(data)
	case "null":
		result = applyNull(data)
	case "t":
		result = applyT(data)
	default:
		result = data
	}
	return
}

// for insertion in a Javscript single or double quoted string
// if data is an UTF-8 string, than the result of "applyJS(data)"" is the same string after evaluation in JS
func applyJS(data string) (result string) {
	if data == "" {
		return
	}
	safe := " 1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ$-+_.!*(),@#{}|[]:;?=^~"

	var buffer strings.Builder
	for _, runeValue := range data {
		if strings.ContainsRune(safe, runeValue) {
			buffer.WriteRune(runeValue)
			continue
		}
		if runeValue < 127 {
			buffer.WriteString("\\x")
			buffer.WriteString(fmt.Sprintf("%02x", runeValue))
			continue
		}
		buffer.WriteString(strconv.QuoteRuneToASCII(runeValue))
	}
	result = buffer.String()
	return
}

func applyPY(data string) (result string) {
	if data == "" {
		return
	}
	safe := " 1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ$-+_.!*(),|[]:;?=^~"

	var buffer strings.Builder
	for _, runeValue := range data {
		if strings.ContainsRune(safe, runeValue) {
			buffer.WriteRune(runeValue)
			continue
		}
		if runeValue < 127 {
			buffer.WriteString("\\x")
			buffer.WriteString(fmt.Sprintf("%02x", runeValue))
			continue
		}
		buffer.WriteString(strconv.QuoteRuneToASCII(runeValue))
	}
	result = buffer.String()
	return
}

func applyPHP(data string) (result string) {
	if data == "" {
		return
	}
	safe := " 1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-+_.!*(),|[]:;?=^~"

	var buffer strings.Builder
	for _, runeValue := range data {
		if strings.ContainsRune(safe, runeValue) {
			buffer.WriteRune(runeValue)
			continue
		}
		if runeValue < 127 {
			buffer.WriteString("\\x")
			buffer.WriteString(fmt.Sprintf("%02x", runeValue))
			continue
		}
		buffer.WriteString(strconv.QuoteRuneToASCII(runeValue))
	}
	result = buffer.String()
	return
}

func applyXML(data string) (result string) {
	if data == "" {
		return
	}
	safe := " 1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-+_.!*(),|[]:;?=^~"

	var buffer strings.Builder
	for _, runeValue := range data {
		if strings.ContainsRune(safe, runeValue) {
			buffer.WriteRune(runeValue)
			continue
		}
		buffer.WriteString("&#")
		buffer.WriteString(fmt.Sprintf("%x;", runeValue))
	}
	result = buffer.String()
	return
}

func applyNull(data string) string {
	return data
}

func applyT(data string) string {
	for _, ch := range `\{}$@%|` {
		source := string(ch)
		target := `\` + source
		data = strings.Replace(data, source, target, -1)
	}
	return data
}
