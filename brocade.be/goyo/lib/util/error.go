package util

import (
	"fmt"

	qprompt "github.com/manifoldco/promptui"
)

func Error(err error) {
	IconBad := qprompt.Styler(qprompt.FGRed)("✗")
	fmt.Println(IconBad, "ERROR:", err)
}
