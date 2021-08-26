package cmd

import (
	"bufio"
	"fmt"
	"os"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceBaseCmd = &cobra.Command{
	Use:     "base",
	Short:   "Return basename",
	Long:    `This comand diplays the basename of a source file`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng source base /catalografie/application/bcawedit.m`,
	RunE:    sourceBase,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	sourceCmd.AddCommand(sourceBaseCmd)
}

func sourceBase(cmd *cobra.Command, args []string) error {

	qpath := qutil.Canon(args[0])
	_, base := qutil.QPartition(qpath)

	if Fstdout == "" || Ftransported {
		fmt.Print(base)
		return nil
	}
	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, base)
	err = w.Flush()
	return err
}
