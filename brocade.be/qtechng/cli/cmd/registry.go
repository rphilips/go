package cmd

import (
	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:     "registry",
	Short:   "registry functions",
	Long:    `All kinds of actions on the registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng registry",
}

func init() {
	registryCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	rootCmd.AddCommand(registryCmd)
}
