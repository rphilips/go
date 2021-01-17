package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var fsReplaceCmd = &cobra.Command{
	Use:     "replace",
	Short:   "replaces stdin to a file",
	Long:    `Command which reads stdin and replaces data to a file in the filesystem`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng fs replace cwd=../catalografie`,
	RunE:    fsReplace,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

//Fappend appends to file
var Fappend bool

func init() {
	fsReplaceCmd.Flags().BoolVar(&Fappend, "append", false, "Appends to file")
	fsCmd.AddCommand(fsReplaceCmd)
}

func fsReplace(cmd *cobra.Command, args []string) error {
	result := args[0]
	if !filepath.IsAbs(result) {
		result = filepath.Join(Fcwd, result)
	}

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
