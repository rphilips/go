package action

import (
	"fmt"
)

func Echo(text string) string {
	fmt.Println(text)
	if text != "" {
		return "echo " + text
	} else {
		return "echo"
	}
}
