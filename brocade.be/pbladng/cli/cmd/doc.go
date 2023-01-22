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
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "doc `gopblad`",
	Long:  "doc `gopblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `gopblad doc myfile.pb`,
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
			args = append(args, filepath.Join(Fcwd, "week.pb"))
		} else {
			args = append(args, pfs.FName("workspace/week.pb"))
		}
	}
	source, err := makeHTML(args[0])
	if err != nil {
		return err
	}

	Fdocty = strings.TrimLeft(Fdocty, ".")
	if Fdocty == "" {
		Fdocty = "doc"
	}
	Fdocty = strings.ToLower(Fdocty)
	target := strings.TrimSuffix(source, ".html") + "." + Fdocty
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
	fmt.Println(strings.Join(convert, " "))
	bfs.Rmpath(odttarget)
	vcmd := exec.Command(convert[0], convert[1:]...)
	err = vcmd.Start()
	if err != nil {
		panic(err)
	}
	err = vcmd.Wait()
	if err != nil {
		panic(err)
	}
	pconvert = pregistry.Registry["html-converter-exe"].([]any)
	convert = make([]string, 0)
	source = odttarget

	if Fdocty != "odt" {
		for _, piece := range pconvert {
			p := piece.(string)
			p = strings.ReplaceAll(p, "{docty}", Fdocty)
			p = strings.ReplaceAll(p, "{outdir}", outdir)
			p = strings.ReplaceAll(p, "{source}", odttarget)
			p = strings.ReplaceAll(p, "{target}", target)
			convert = append(convert, p)
		}
		//fmt.Println(strings.Join(convert, " "))
		bfs.Rmpath(target)
		vcmd := exec.Command(convert[0], convert[1:]...)
		err = vcmd.Start()
		if err != nil {
			panic(err)
		}
		err = vcmd.Wait()
		if err != nil {
			panic(err)
		}
	}

	//fmt.Println(target)

	pviewer := pregistry.Registry["viewer"].(map[string]any)[Fdocty].([]any)
	viewer := make([]string, 0)

	for _, piece := range pviewer {
		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", target))
	}
	vcmd = exec.Command(viewer[0], viewer[1:]...)
	vcmd.Stderr = io.Discard
	vcmd.Stdout = io.Discard
	err = vcmd.Start()
	if err != nil {
		panic(err)
	}

	return err
}
