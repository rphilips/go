package tools

import "fmt"

func Error(code string, lineno int, msg string) error {
	return fmt.Errorf("ERROR %s: line %d: %s", code, lineno, msg)
}
