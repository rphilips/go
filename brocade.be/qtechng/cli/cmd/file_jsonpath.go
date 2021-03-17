package cmd

import (
	"fmt"
	"io"
	"os"

	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
	"github.com/spyzhov/ajson"
)

var fileJsonpathCmd = &cobra.Command{
	Use:     "jsonpath",
	Short:   "jsonpath selection",
	Long:    `Filters stdin - as a JSON string - through jsonpath and writes on stdout`,
	Example: "  qtechng file jsonpath '$.store.book[*].author'",
	Args:    cobra.ExactArgs(1),
	RunE:    fileJsonpath,
}

func init() {
	fileCmd.AddCommand(fileJsonpathCmd)
}

func fileJsonpath(cmd *cobra.Command, args []string) (err error) {
	jsonpath := args[0]
	_, err = ajson.ParseJSONPath(jsonpath)
	if err != nil {
		err = &qerror.QError{
			Ref: []string{errRoot + "jsonpath"},
			Msg: []string{fmt.Sprintf("JSONpath `" + jsonpath + "` error: " + err.Error())},
		}
		return
	}
	data, err := io.ReadAll(os.Stdin)
	result, err := ajson.JSONPath(data, jsonpath)

	if err != nil {
		return err
	}

	if Fstdout == "" || Ftransported {
		fmt.Print(result)
		return nil
	}

	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprint(f, result)
	return err
}
