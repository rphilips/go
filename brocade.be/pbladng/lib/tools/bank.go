package tools

import (
	"regexp"
	"strings"

	"github.com/jbub/banking/iban"
)

var rbank = regexp.MustCompile(`\b[Bb][Ee][0123456789 -]{13,}[0123456789]\b`)

func CheckIBAN(s string) error {
	parts := rbank.FindAllStringIndex(s, -1)
	if parts == nil {
		return nil
	}
	for i := 0; i < len(parts); i++ {
		bank := s[parts[i][0]:parts[i][1]]
		bank = strings.ReplaceAll(bank, " ", "")
		bank = strings.ReplaceAll(bank, "-", "")
		bank = strings.ToUpper(bank)
		err := iban.Validate(bank)
		if err != nil {
			return err
		}
	}
	return nil
}

func Bank(s string, bolden bool) string {
	parts := rbank.FindAllStringIndex(s, -1)
	if parts == nil {
		return s
	}
	result := ""
	if parts[0][0] != 0 {
		result += s[0:parts[0][0]]
	}
	for i := 0; i < len(parts); i++ {
		bank := s[parts[i][0]:parts[i][1]]
		after := ""
		if i+1 < len(parts) {
			after = s[parts[i][1]:parts[i+1][0]]
		} else {
			after = s[parts[i][1]:]
		}
		x := strings.ReplaceAll(bank, " ", "")
		x = strings.ReplaceAll(x, "-", "")
		if len(x) != 16 {
			result += bank + after
			continue
		}
		bank = strings.ToUpper(x)
		bank = bank[:4] + "\u00A0" + bank[4:8] + "\u00A0" + bank[8:12] + "\u00A0" + bank[12:]

		if !bolden {
			result += bank + "\x00" + after
			continue
		}

		x = strings.ReplaceAll(result, `\\`, "\x02")
		x = strings.ReplaceAll(x, `\*`, "\x04")
		if strings.Count(x, "*")%2 == 0 {
			result += "*" + bank + "*\x00" + after
		} else {
			result += bank + "\x00" + after
		}
	}
	result = strings.TrimRight(result, "\x00")
	result = strings.ReplaceAll(result, "\x00 .", ".")
	result = strings.ReplaceAll(result, "\x00 ", " ")
	result = strings.ReplaceAll(result, "\x00", "")
	result = strings.ReplaceAll(result, "...", "\x00")
	result = strings.ReplaceAll(result, "..", ".")
	result = strings.ReplaceAll(result, ". .", ".")
	result = strings.ReplaceAll(result, "\x00", "...")
	return result
}
