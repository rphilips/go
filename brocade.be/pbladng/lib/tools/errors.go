package tools

import "fmt"

func Error(code string, lineno int, msg any) error {
	mesg := ""
	switch v := msg.(type) {
	case string:
		mesg = v
	case fmt.Stringer:
		mesg = v.String()
	case error:
		mesg = v.Error()
	default:
		mesg = fmt.Sprintf("%v", msg)
	}
	return fmt.Errorf("ERROR %s: line %d: %s", code, lineno, mesg)
}
