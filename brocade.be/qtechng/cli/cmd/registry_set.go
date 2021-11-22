package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	qregistry "brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var registrySetCmd = &cobra.Command{
	Use:   "set [key [value]]",
	Short: "Set registry keys",
	Long: `With two arguments, this command sets a registry key and value.
Without arguments, it interactively asks for the values`,
	Example: `
  qtechng registry set scratch-dir /home/rphilips/tmp
  qtechng registry set`,

	Args: cobra.RangeArgs(1, 2),
	RunE: registrySet,
	Annotations: map[string]string{
		"remote-allowed":       "no",
		"rstrip-trailing-crlf": "yes",
	},
}

func init() {
	registryCmd.AddCommand(registrySetCmd)
}

func registrySet(cmd *cobra.Command, args []string) (err error) {

	key := args[0]
	value := ""
	ok := false
	switch len(args) {
	case 2:
		value = args[1]
	case 1:
		reader := bufio.NewReader(os.Stdin)
		value, ok = qregistry.Registry[key]
		prompt := "Value (empty is NOT a value): "
		if ok {
			fmt.Println(key, "->", value)
			prompt = "New value (empty is NOT a value): "
		}
		fmt.Print(prompt)
		value, _ = reader.ReadString('\n')
		value = strings.TrimSuffix(value, "\n")
	}
	if value != "" {
		oldvalue, ok := qregistry.Registry[key]
		if !ok || oldvalue != value {
			qregistry.SetRegistry(key, value)
			qregistry.Registry[key] = value
		}
	}

	return nil
}
