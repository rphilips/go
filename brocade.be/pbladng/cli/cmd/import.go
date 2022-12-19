package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pdocument "brocade.be/pbladng/lib/document"
	pregistry "brocade.be/pbladng/lib/registry"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import documents from the correspondents directories",
	Long:  "import documents from the correspondents directories",

	Args:    cobra.NoArgs,
	Example: `gopblad import`,
	RunE:    doimport,
}

func init() {

	rootCmd.AddCommand(importCmd)
}

func doimport(cmd *cobra.Command, args []string) error {
	install(cmd, args)
	err := pdocument.Archive(Fcwd)
	if err != nil {
		return err
	}
	year, week, _ := pdocument.DocRef(Fcwd)
	dst := Fcwd

	correspondents := pregistry.Registry["correspondents"].(map[string]any)
	for prefix, dir := range correspondents {
		src := filepath.Join(dir.(string), fmt.Sprintf("%d", year), fmt.Sprintf("%02d", week))
		if !bfs.IsDir(src) {
			fmt.Println("No files found for", prefix)
			continue
		}
		//src = pfs.FName(src)
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		ok := false
		for _, ery := range entries {
			entry, e := ery.Info()
			if e != nil {
				continue
			}
			if entry.IsDir() {
				continue
			}
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}
			name := entry.Name()
			if name == "week.md" {
				continue
			}
			srcPath := filepath.Join(src, name)
			ext := path.Ext(name)
			root := strings.TrimSuffix(name, ext)
			ext = strings.ToLower(ext)
			// new name: prefix-root.ext
			dstPath := filepath.Join(dst, prefix+"-"+root+ext)
			err = bfs.CopyFile(srcPath, dstPath, "", false)
			if err != nil {
				return err
			}
			ok = true
		}
		if !ok {
			fmt.Println("No files found for", prefix)
			continue
		}
	}
	return nil
}
