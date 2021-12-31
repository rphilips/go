package cmd

import (
	"io"
	"os"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsStoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "Store stdin to a file",
	Long:    `This commands reads stdin and stores data to a file in the filesystem`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng fs store receive.txt --cwd=../workspace`,
	RunE:    fsStore,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsStoreCmd.Flags().BoolVar(&Fappend, "append", false, "Appends to file")
	fsCmd.AddCommand(fsStoreCmd)
}

func fsStore(cmd *cobra.Command, args []string) error {
	result := qutil.AbsPath(args[0], Fcwd)

	mode := os.O_WRONLY | os.O_CREATE
	if Fappend {
		mode = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	}

	f, err := os.OpenFile(result, mode, 0644)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	_, err = io.Copy(f, os.Stdin)
	if err != nil {
		panic(err)
	}
	return err
}
