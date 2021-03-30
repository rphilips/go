package cmd

import (
	"bufio"
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

// Fmdb directory with the M database
var Fmdb string

// Fbulk send to M in bulk
var Fbulk bool

func init() {
	stdinMumpsCmd.Flags().StringVar(&Fmdb, "mdb", "", "directory with the M database")
	stdinMumpsCmd.Flags().BoolVar(&Fbulk, "bulk", false, "send to M in bulk")
	stdinCmd.AddCommand(stdinMumpsCmd)
}

func stdinMumps(cmd *cobra.Command, args []string) (err error) {
	if Fbulk {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		buffer := bytes.NewBuffer(data)
		err = qmumps.PipeTo(Fmdb, []*bytes.Buffer{buffer})

		return err
	}
	// send line per line
	reader := bufio.NewReader(os.Stdin)
	err = qmumps.PipeLineTo(Fmdb, reader)
	return err
}
