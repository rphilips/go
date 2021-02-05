package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var stdinJSONCmd = &cobra.Command{
	Use:     "json",
	Short:   "Beautifies JSON input",
	Long:    `Receives JSON on stdin, beautifies it and writes to stdout. Subscripts are sorted.`,
	Example: `qtechng stdin json '{"b":"B", "a":"A"}'`,
	Args:    cobra.NoArgs,
	RunE:    stdinJSON,
}

func init() {
	stdinCmd.AddCommand(stdinJSONCmd)
}

func stdinJSON(cmd *cobra.Command, args []string) (err error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	var ifce interface{}
	err = json.Unmarshal(data, &ifce)
	if err != nil {
		return err
	}
	data, err = json.Marshal(ifce)
	if err != nil {
		return err
	}

	var out bytes.Buffer

	err = json.Indent(&out, data, "", "    ")
	if err == nil {
		if Fstdout == "" {
			_, err := fmt.Println(out.String())
			return err
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		_, err = fmt.Fprintln(f, out.String())
		return err
	}
	return err
}
