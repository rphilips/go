package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var fsStoreCmd = &cobra.Command{
	Use:     "store",
	Short:   "stores stdin to a file",
	Long:    `Command which reads stdin and stores to a file in the filesystem`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng fs store cwd=../catalografie`,
	RunE:    fsStore,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

//Fappend appends to file
var Fappend bool

func init() {
	fsStoreCmd.Flags().BoolVar(&Fappend, "append", false, "Appends to file")
	fsCmd.AddCommand(fsStoreCmd)
}

func fsStore(cmd *cobra.Command, args []string) error {
	result := filepath.Join(Fcwd, args[0])

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
