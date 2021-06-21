package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"

	qfnmatch "brocade.be/base/fnmatch"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var registryGetCmd = &cobra.Command{
	Use:   "get pattern",
	Short: "Retrieves the registry values",
	Long: `
List all registry values, with the key matching a pattern, and writes on stdout`,
	Example: `
  qtechng registry get scratch-dir
  qtechng registry get qtechng-*`,

	Args:   cobra.ExactArgs(1),
	RunE:   registryGet,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed":       "yes",
		"rstrip-trailing-crlf": "yes",
	},
}

func init() {
	registryCmd.AddCommand(registryGetCmd)
}

func registryGet(cmd *cobra.Command, args []string) (err error) {

	pattern := args[0]
	found := make([]string, 0)
	for key := range qregistry.Registry {
		if qfnmatch.Match(pattern, key) {
			found = append(found, key)
		}
	}
	sort.Strings(found)
	if len(found) == 1 {
		if Fstdout == "" || Ftransported {
			fmt.Printf("%s", qregistry.Registry[found[0]])
			return nil
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fmt.Fprintf(f, "%s", qregistry.Registry[found[0]])
	}

	msg := make(map[string]string)
	for _, key := range found {
		msg[key] = qregistry.Registry[key]
	}

	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
