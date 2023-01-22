package cmd

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
)

var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "HTML `gopblad`",
	Long:  "HTML `gopblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `gopblad html myfile.pb`,
	RunE:    HTML,
}

func init() {

	rootCmd.AddCommand(htmlCmd)
}

func HTML(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if Fdebug {
			Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
			args = append(args, filepath.Join(Fcwd, "week.pb"))
		} else {
			args = append(args, pfs.FName("workspace/week.pb"))
		}
	}
	target, err := makeHTML(args[0])
	if err != nil {
		return err
	}

	pviewer := pregistry.Registry["viewer"].(map[string]any)["html"].([]any)
	viewer := make([]string, 0)

	for _, piece := range pviewer {
		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", target))
	}
	vcmd := exec.Command(viewer[0], viewer[1:]...)
	err = vcmd.Start()
	if err != nil {
		panic(err)
	}

	return err
}

func makeHTML(fname string) (target string, err error) {

	var source io.Reader
	dir := pfs.FName("workspace")
	if fname == "-" {
		source = os.Stdin
	} else {
		file, e := os.Open(fname)
		err = e
		if err != nil {
			return
		}
		dir = filepath.Dir(fname)
		source = bufio.NewReader(file)
	}
	doc := new(pstructure.Document)
	doc.Dir = dir
	err = doc.Load(source)
	if err != nil {
		return
	}
	target = filepath.Join(dir, "parochieblad.html")
	bfs.Store(target, doc.HTML(), "process")
	return
}
