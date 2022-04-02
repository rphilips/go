package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var stdinJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Beautify JSON input",
	Long: `This command receives JSON on stdin, beautifies it and writes to stdout. Subscripts are sorted.
You can also provide the JSON as an argument`,
	Example: `echo '{"b":"B", "a":"A"}' | qtechng stdin json
qtechng stdin json '{"b":"B", "a":"A"}'`,
	Args: cobra.MaximumNArgs(1),
	RunE: stdinJSON,
}

func init() {
	stdinCmd.AddCommand(stdinJSONCmd)
}

func stdinJSON(cmd *cobra.Command, args []string) (err error) {
	var reader *bufio.Reader
	if len(args) == 0 {
		reader = bufio.NewReader(os.Stdin)
	} else {
		reader = bufio.NewReader(strings.NewReader(args[0]))
	}

	data, err := io.ReadAll(reader)
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
