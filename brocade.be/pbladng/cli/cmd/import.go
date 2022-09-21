package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
	pmanuscript "brocade.be/pbladng/lib/manuscript"
	pregistry "brocade.be/pbladng/lib/registry"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import `gopblad`",
	Long:  "import `gopblad`",

	Args:    cobra.NoArgs,
	Example: `gopblad import`,
	RunE:    doimport,
}

func init() {

	rootCmd.AddCommand(importCmd)
}

func doimport(cmd *cobra.Command, args []string) error {
	install(cmd, args)
	weekpb := pfs.FName("workspace/week.pb")
	f, err := os.Open(weekpb)
	if err != nil {
		return err
	}
	defer f.Close()
	source := bufio.NewReader(f)
	m, err := pmanuscript.Parse(source, false, "")
	if err != nil {
		return err
	}
	year := m.Year
	week := m.Week

	wdir := fmt.Sprintf("%d/%02d", year, week)
	dst := pfs.FName("workspace")

	correspondents := pregistry.Registry["correspondents"].(map[string]string)
	for prefix, dir := range correspondents {
		src := dir + "/" + wdir
		src = pfs.FName(src)
		entries, err := os.ReadDir(src)
		if err != nil {
			continue
		}
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
			if name == "week.pb" {
				continue
			}
			srcPath := filepath.Join(src, name)
			ext := path.Ext(name)
			root := name
			if ext != "" {
				root = strings.TrimSuffix(name, ext)
			}
			if root == "" {
				ext = "." + prefix + "-" + ext[1:]
			} else {
				root = root + "-" + prefix
			}
			dstPath := filepath.Join(dst, root+ext)

			err = bfs.CopyFile(srcPath, dstPath, "process", false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
