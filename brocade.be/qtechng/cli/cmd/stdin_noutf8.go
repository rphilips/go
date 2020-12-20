package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

var stdinNoUTF8Cmd = &cobra.Command{
	Use:     "noutf8",
	Short:   "Searches for non UTF-8 characters",
	Long:    "Searches stdin per line for first non UTF-8 character.\nwrites a JSON report on stdout",
	Example: "qtechng stdin noutf8",
	Args:    cobra.NoArgs,
	RunE:    stdinNoUTF8,
}

func init() {
	stdinCmd.AddCommand(stdinNoUTF8Cmd)
}

func stdinNoUTF8(cmd *cobra.Command, args []string) (err error) {
	repl := rune(65533)
	result := [][2]int{}

	reader := bufio.NewReader(os.Stdin)

	count := 0
	for {
		count++
		line, err := reader.ReadSlice(10)
		if err != nil && err != io.EOF {
			break
		}
		if utf8.Valid(line) && !bytes.ContainsRune(line, repl) {
			if err == io.EOF {
				break
			}
			continue
		}

		good := strings.ToValidUTF8(string(line), string(repl))
		parts := strings.SplitN(good, string(repl), -1)

		total := ""
		for c, part := range parts {
			if c == len(parts) {
				break
			}
			total += part + "\n"
			result = append(result, [2]int{count, len([]rune(total))})
		}
	}
	if len(result) == 0 {
		if Fstdout == "" || Ftransported {
			fmt.Println("[]")
			return nil
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		fmt.Println("[]")
		err = w.Flush()
		return err
	}

	j, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if Fstdout == "" || Ftransported {
		fmt.Println(j)
		return nil
	}

	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprintln(w, j)
	err = w.Flush()
	return err
}
