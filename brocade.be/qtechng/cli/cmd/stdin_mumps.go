package cmd

import (
	"bytes"
	"io"
	"os"

	"github.com/spf13/cobra"

	qmumps "brocade.be/base/mumps"
)

var stdinMumpsCmd = &cobra.Command{
	Use:   "mumps",
	Short: "sends stdin to M",
	Long:  `the lines, read from stdin, are M commands and they are sent to M`,
	Example: `
  qtechng stdin mumps`,
	Args: cobra.NoArgs,
	RunE: stdinMumps,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "B",
	},
}

var Fmdb string

func init() {
	stdinMumpsCmd.Flags().StringVar(&Fmdb, "mdb", "", "directory with the M database")
	stdinCmd.AddCommand(stdinMumpsCmd)
}

func stdinMumps(cmd *cobra.Command, args []string) (err error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(data)
	err = qmumps.PipeTo(Fmdb, []*bytes.Buffer{buffer})

	return err
}
