package cmd

import (
	"fmt"

	"github.com/abiosoft/ishell/v2"
)

func greet(c *ishell.Context) {
	text := AboutText()
	if c == nil {
		fmt.Println(text)
	} else {
		c.Println(text)
	}
}
