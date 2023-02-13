package cmd

import (
	"path/filepath"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pregistry "brocade.be/pbladng/lib/registry"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install `pblad`",
	Long:  "Install `pblad`",

	Args:    cobra.NoArgs,
	Example: `pblad install`,
	RunE:    install,
}

func init() {

	rootCmd.AddCommand(installCmd)
}

func install(cmd *cobra.Command, args []string) error {
	basedir := pregistry.Registry["base-dir"].(string)
	bfs.MkdirAll(basedir, "process")
	for _, sub := range []string{"workspace", "archive/manuscripts", "archive/subscriptions"} {
		sub := filepath.FromSlash(sub)
		bfs.MkdirAll(filepath.Join(basedir, sub), "process")
	}

	correspondents := pregistry.Registry["correspondents"].(map[string]any)
	for _, info := range correspondents {
		bfs.MkdirAll(filepath.Join(basedir, info.(map[string]any)["dir"].(string)), "process")
	}

	return nil
}
