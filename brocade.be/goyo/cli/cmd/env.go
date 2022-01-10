package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:     "env",
	Short:   "setting environment variables",
	Long:    `setting environment variables`,
	Args:    cobra.RangeArgs(1, 2),
	Example: "goya env GOYA_DIR /tmp/abc",
	RunE:    setenv,
}

func init() {
	rootCmd.AddCommand(envCmd)
}

func setenv(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := ""
	if len(args) > 1 {
		value = args[1]
	}
	if key == "" {
		return nil
	}
	return os.Setenv(key, value)
}
