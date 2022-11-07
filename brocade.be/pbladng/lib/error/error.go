package error

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

func Markdown(name string, lno int, incode int, inbold int, initalic int) error {
	if incode > -1 {
		err := Error("doc-"+name+"-incode", lno, fmt.Sprintf("%s in codespan starting at line %d", name, incode))
		return err
	}
	if inbold > -1 {
		err := Error("doc-"+name+"-incode", lno, fmt.Sprintf("%s in codespan starting at line %d", name, incode))
		return err
	}
	if initalic > -1 {
		err := Error("doc-"+name+"-incode", lno, fmt.Sprintf("%s in codespan starting at line %d", name, incode))
		return err
	}
	return nil
}
