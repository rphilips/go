package cmd

import (
	"bufio"
	"os"

	qmumps "brocade.be/base/mumps"
	"github.com/spf13/cobra"
)

var tomumpsCmd = &cobra.Command{
	Use:     "tomumps",
	Short:   "sends M statements from stdin to M",
	Long:    "sends M statements from stdin to M",
	Args:    cobra.NoArgs,
	Example: `goyo tomumps`,
	RunE:    tomumps,
}

var Fmdb string

func init() {
	tomumpsCmd.Flags().StringVar(&Fmdb, "mdb", "", "Directory in the M database")
	rootCmd.AddCommand(tomumpsCmd)

}

func tomumps(cmd *cobra.Command, args []string) error {
	mpipe, err := qmumps.Open(Fmdb)
	if err != nil {
		return err
	}
	defer mpipe.Close()
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		mpipe.WriteExec(scanner.Text())
	}
	return nil
}
