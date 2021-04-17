package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os/exec"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
	"github.com/zserge/lorca"
)

// Finplace replace the file contents

var guiExampleCmd = &cobra.Command{
	Use:     "example",
	Short:   "GUI example",
	Long:    `An example of the use of a GUI`,
	Args:    cobra.NoArgs,
	RunE:    guiExample,
	Example: "qtechng gui example",
}

func init() {
	guiCmd.AddCommand(guiExampleCmd)
}

func guiExample(cmd *cobra.Command, args []string) error {
	// Create UI with basic HTML passed via data URI

	t, err := template.ParseFS(guifs, "templates/example.html")
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, guiFiller)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}

	ui, err := lorca.New("data:text/html,"+url.PathEscape(buf.String()), "", 480, 320)
	if err != nil {
		log.Fatal(err)
	}
	ui.Bind("golangfunc", func() {
		qpattern := ui.Eval(`document.getElementById('qpattern').value`)
		version := ui.Eval(`document.getElementById('version').value`)

		if qpattern.String() != "" && version.String() != "" {
			qp := qpattern.String()
			vs := version.String()
			fmt.Println("qp:", qp)
			fmt.Println("vs:", vs)
			var stdout bytes.Buffer
			var stderr bytes.Buffer
			argums := []string{
				qregistry.Registry["qtechng-exe"],
				"source",
				"list",
				"--version=" + vs,
				"--qpattern=" + qp,
				"--jsonpath=$..qpath",
				"--yaml",
			}
			qexe, _ := exec.LookPath(qregistry.Registry["qtechng-exe"])
			cmd := exec.Cmd{
				Path:   qexe,
				Args:   argums,
				Dir:    qregistry.Registry["scratch-dir"],
				Stdout: &stdout,
				Stderr: &stderr,
			}
			err := cmd.Run()
			sout := stdout.String()
			fmt.Println(sout)
			serr := stderr.String()
			fmt.Println(serr)
			if err != nil {
				serr += "\n\nError:" + err.Error()
			}
			bx, _ := json.Marshal(sout)

			ui.Eval(`document.getElementById("response").innerHTML = ` + string(bx))

		}
	})
	defer ui.Close()
	// Wait until UI window is closed
	<-ui.Done()
	return nil
}
