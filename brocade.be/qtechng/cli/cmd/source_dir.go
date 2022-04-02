package cmd

import (
	"bufio"
	"fmt"
	"os"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceDirCmd = &cobra.Command{
	Use:     "dir",
	Short:   "Return dirname",
	Long:    `Command to display the directory name of a source file`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng source dir /catalografie/application/bcawedit.m`,
	RunE:    sourceDir,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	sourceCmd.AddCommand(sourceDirCmd)
}

func sourceDir(cmd *cobra.Command, args []string) error {

	qpath := qutil.Canon(args[0])
	dir, _ := qutil.QPartition(qpath)

	if Fstdout == "" || Ftransported {
		fmt.Print(dir)
		return nil
	}
	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, dir)
	err = w.Flush()
	return err
}
