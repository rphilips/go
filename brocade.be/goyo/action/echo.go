package action

import (
	"fmt"
)

func Echo(text string) []string {
	fmt.Println(text)
	if text != "" {
		return []string{"echo " + text}
	} else {
		return []string{"echo"}
	}
}
