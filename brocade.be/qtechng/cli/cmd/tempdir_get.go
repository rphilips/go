package cmd

import (
	"bufio"
	"fmt"
	"os"

	qfs "brocade.be/base/fs"
	"github.com/spf13/cobra"
)

// Fprefix is een prefix voor tijdelijke directory
var Fprefix = ""

var tempdirGetCmd = &cobra.Command{
	Use:   "get",
	Short: "creates a temporary directory",
	Long: `
Creates and returns a temporary directory`,
	Example: "  qtechng tempdir get\n  qtechng tempdir get --prefix=qtechng.",
	Args:    cobra.NoArgs,
	RunE:    tempdirGet,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
	},
}

func init() {
	tempdirGetCmd.Flags().StringVar(&Fprefix, "prefix", "", "prefix to append")
	tempdirGetCmd.Flags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	tempdirCmd.AddCommand(tempdirGetCmd)
}

func tempdirGet(cmd *cobra.Command, args []string) error {

	tempdir, err := qfs.TempDir("", Fprefix)
	if err != nil {
		return err
	}
	if Fstdout == "" || Ftransported {
		fmt.Print(tempdir)
		return nil
	}
	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, tempdir)
	err = w.Flush()
	return err
}
