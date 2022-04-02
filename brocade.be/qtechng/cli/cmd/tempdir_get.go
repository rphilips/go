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
	Short: "Create a temporary directory",
	Long:  `This command creates and returns a temporary directory`,
	Example: `qtechng tempdir get
qtechng tempdir get --prefix=qtechng.`,
	Args:   cobra.NoArgs,
	RunE:   tempdirGet,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
	},
}

func init() {
	tempdirGetCmd.Flags().StringVar(&Fprefix, "prefix", "", "Prefix to append to the tempdir name")
	tempdirGetCmd.Flags().BoolVar(&Fremote, "remote", false, "Execute on the remote server")
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
