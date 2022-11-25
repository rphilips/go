package strings

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func UTF8(buf []byte) string {
	if len(buf) == 0 {
		return ""
	}
	if utf8.Valid(buf) {
		return string(buf)
	}
	bufr := make([]rune, len(buf))
	for i, b := range buf {
		bufr[i] = rune(b)
	}
	return string(bufr)
}

func YesNo(s string) bool {
	for {
		fmt.Printf("%s [y/n] ", s)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		if strings.HasPrefix(text, "y") {
			return true
		}
		if strings.HasPrefix(text, "n") {
			return false
		}
	}
}
