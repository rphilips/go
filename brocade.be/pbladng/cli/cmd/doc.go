package cmd

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "doc `pblad`",
	Long:  "doc `pblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `pblad doc myfile.pb`,
	RunE:    doc,
}

var Fdocty = ""

func init() {
	docCmd.PersistentFlags().StringVar(&Fdocty, "docty", "", "document type: doc, docx, pdf")
	rootCmd.AddCommand(docCmd)
}

func doc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if Fdebug {
			Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
			args = append(args, filepath.Join(Fcwd, "parochieblad.ed"))
		} else {
			args = append(args, pfs.FName("workspace/parochieblad.ed"))
		}
	}

	if Fdocty == "" {
		Fdocty = "doc,docx"
	}
	_, target, err := makeDoc(args[0], Fdocty)
	if err != nil {
		return err
	}

	fmt.Println(target)
	Fdocty, _, _ = strings.Cut(Fdocty, ",")
	target, _, _ = strings.Cut(target, ",")

	pviewer := pregistry.Registry["viewer"].(map[string]any)[Fdocty].([]any)
	viewer := make([]string, 0)

	for _, piece := range pviewer {
		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", target))
	}
	vcmd := exec.Command(viewer[0], viewer[1:]...)
	vcmd.Stderr = io.Discard
	vcmd.Stdout = io.Discard
	err = vcmd.Start()
	return err
}

func makeDoc(file string, docty string) (doc *pstructure.Document, targets string, err error) {

	doc, source, err := makeHTML(file)
	if err != nil {
		return
	}

	odttarget := strings.TrimSuffix(source, ".html") + ".odt"
	outdir := filepath.Dir(source)
	if outdir == "" {
		outdir = "."
	}
	pconvert := pregistry.Registry["html-converter-exe"].([]any)
	convert := make([]string, 0)

	for _, piece := range pconvert {
		p := piece.(string)
		p = strings.ReplaceAll(p, "{docty}", "odt")
		p = strings.ReplaceAll(p, "{outdir}", outdir)
		p = strings.ReplaceAll(p, "{source}", source)
		p = strings.ReplaceAll(p, "{target}", odttarget)
		convert = append(convert, p)
	}
	bfs.Rmpath(odttarget)
	vcmd := exec.Command(convert[0], convert[1:]...)
	err = vcmd.Start()
	if err != nil {
		return
	}
	err = vcmd.Wait()
	if err != nil {
		return
	}

	doctys := strings.SplitN(docty, ",", -1)

	targets = ""
	for _, docty := range doctys {
		docty = strings.TrimLeft(strings.TrimSpace(docty), ".")
		if docty == "" {
			continue
		}
		docty = strings.ToLower(docty)
		target := strings.TrimSuffix(source, ".html") + "." + docty
		if docty == "odt" {
			targets += "," + target
			targets = strings.TrimPrefix(targets, ",")
			continue
		}

		convert := make([]string, 0)

		for _, piece := range pconvert {
			p := piece.(string)
			p = strings.ReplaceAll(p, "{docty}", docty)
			p = strings.ReplaceAll(p, "{outdir}", outdir)
			p = strings.ReplaceAll(p, "{source}", odttarget)
			p = strings.ReplaceAll(p, "{target}", target)
			convert = append(convert, p)
		}
		bfs.Rmpath(target)
		vcmd := exec.Command(convert[0], convert[1:]...)
		err = vcmd.Start()
		if err != nil {
			continue
		}
		err = vcmd.Wait()
		if err != nil {
			continue
		}
		targets += "," + target
		targets = strings.TrimPrefix(targets, ",")
	}

	return
}
