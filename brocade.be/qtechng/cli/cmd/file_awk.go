package cmd

import (
	"log"
	"os"

	qawk "github.com/benhoyt/goawk/interp"
	"github.com/spf13/cobra"
)

var fileAwkCmd = &cobra.Command{
	Use:   "awk statement",
	Short: "AWK interpreter",
	Long: `
Filters stdin through through an AWK statement and writes on stdout`,
	Example: `
  qtechng file awk '{print $2}'`,

	Args: cobra.ExactArgs(1),
	RunE: fileAwk,
}

func init() {
	fileCmd.AddCommand(fileAwkCmd)
}

func fileAwk(cmd *cobra.Command, args []string) (err error) {
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
