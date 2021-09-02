package mumps

import (
	"github.com/fatih/color"
)

var Ask = color.New(color.FgGreen).SprintFunc()
var Info = color.New(color.FgBlue).SprintFunc()
var Error = color.New(color.FgRed).SprintFunc()
