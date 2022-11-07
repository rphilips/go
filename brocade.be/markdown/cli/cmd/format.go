package cmd

import (
	"fmt"
	"os"

	mparse "brocade.be/markdown/lib/parse"
	"github.com/spf13/cobra"
	mast "github.com/yuin/goldmark/ast"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "format `markdown`",
	Long:  "format `markdown`",

	Example: `markdown format myfile.pb`,
	RunE:    format,
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

func format(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, "/home/rphilips/go/brocade.be/markdown/test/week.md")
	}
	fname := args[0]
	blob, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	doc := mparse.Parse(blob)
	Dump(doc, blob, "", 0)

	return nil
}

func Dump(v mast.Node, source []byte, pname string, start int) {

	name := v.Kind().String()
	ty := v.Type()
	switch ty {
	case mast.TypeBlock:
		line := v.Lines().At(0)
		start = line.Start
		fmt.Println("NAME:", name, start)
	case mast.TypeInline:

		value := string(v.Text(source))

		// if v.SoftLineBreak() {
		// 	value += "\n"
		// }
		fmt.Println(name, "=", value)

	}

	for c := v.FirstChild(); c != nil; c = c.NextSibling() {
		Dump(c, source, name, start)
	}
	// value := ""
	// switch ty {
	// case mast.TypeBlock:
	// 	line := v.Lines().At(0)
	// 	start = line.Start
	// default:
	// 	for i := 0; i < v.Lines().Len(); i++ {
	// 		line := v.Lines().At(i)
	// 		value := line.Value(source)
	// 	}
	// }
	// if ty == mast.TypeBlock {
	// 	line := v.Lines().At(0)
	// 	start = line.Start
	// }

	// for c := v.FirstChild(); c != nil; c = c.NextSibling() {
	// 	Dump(c, source, name, start)
	// }

	// fmt.Printf("%s {\n", name)
	// if v.Type() == mast.TypeBlock {
	// 	for i := 0; i < v.Lines().Len(); i++ {
	// 		line := v.Lines().At(i)
	// 		fmt.Printf("[%s...%s] %d %d: %s\n", string(source[line.Start:line.Start+1]), string(source[line.Stop-1:line.Stop]), line.Start, line.Stop, line.Value(source))
	// 	}
	// 	fmt.Printf("\"\n====\n")
	// 	for c := v.FirstChild(); c != nil; c = c.NextSibling() {
	// 		Dump(c, source)
	// 	}
	// 	fmt.Printf("}\n")
	// } else {
	// 	v.Dump(source, 0)
	// }
}
