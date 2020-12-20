package cmd

import (
	"log"
	"os"

	qawk "github.com/benhoyt/goawk/interp"
	"github.com/spf13/cobra"
)

var stdinAwkCmd = &cobra.Command{
	Use:   "awk statement",
	Short: "AWK interpreter",
	Long: `
Filters stdin through through an AWK statement and writes on stdout`,
	Example: `
  qtechng stdin awk '{print $2}'`,

	Args: cobra.ExactArgs(1),
	RunE: stdinAwk,
}

func init() {
	stdinCmd.AddCommand(stdinAwkCmd)
}

func stdinAwk(cmd *cobra.Command, args []string) (err error) {
	if Fstdout == "" || Ftransported {
		err = qawk.Exec(args[0], " ", nil, nil)
		return nil
	}

	f, err := os.Create(Fstdout)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = qawk.Exec(args[0], " ", nil, f)
	return err
}
