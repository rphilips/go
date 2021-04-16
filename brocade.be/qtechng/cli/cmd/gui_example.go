package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os/exec"

	qregistry "brocade.be/base/registry"
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
	ui, err := lorca.New("data:text/html,"+url.PathEscape(`<!DOCTYPE html>
	<html>
		<head><title>QtechNG</title></head>
		<style type="text/css">
.form-style-2{
	max-width: 500px;
	padding: 20px 12px 10px 20px;
	font: 13px Arial, Helvetica, sans-serif;
}
.form-style-2-heading{
	font-weight: bold;
	font-style: italic;
	border-bottom: 2px solid #ddd;
	margin-bottom: 20px;
	font-size: 15px;
	padding-bottom: 3px;
}
.form-style-2 label{
	display: block;
	margin: 0px 0px 15px 0px;
}
.form-style-2 label > span{
	width: 100px;
	font-weight: bold;
	float: left;
	padding-top: 8px;
	padding-right: 5px;
}
.form-style-2 span.required{
	color:red;
}
.form-style-2 .tel-number-field{
	width: 40px;
	text-align: center;
}
.form-style-2 input.input-field, .form-style-2 .select-field{
	width: 48%;	
}
.form-style-2 input.input-field, 
.form-style-2 .tel-number-field, 
.form-style-2 .textarea-field, 
 .form-style-2 .select-field{
	box-sizing: border-box;
	-webkit-box-sizing: border-box;
	-moz-box-sizing: border-box;
	border: 1px solid #C2C2C2;
	box-shadow: 1px 1px 4px #EBEBEB;
	-moz-box-shadow: 1px 1px 4px #EBEBEB;
	-webkit-box-shadow: 1px 1px 4px #EBEBEB;
	border-radius: 3px;
	-webkit-border-radius: 3px;
	-moz-border-radius: 3px;
	padding: 7px;
	outline: none;
}
.form-style-2 .input-field:focus, 
.form-style-2 .tel-number-field:focus, 
.form-style-2 .textarea-field:focus,  
.form-style-2 .select-field:focus{
	border: 1px solid #0C0;
}
.form-style-2 .textarea-field{
	height:100px;
	width: 55%;
}
.form-style-2 input[type=submit],
.form-style-2 input[type=button]{
	border: none;
	padding: 8px 15px 8px 15px;
	background: #FF8500;
	color: #fff;
	box-shadow: 1px 1px 4px #DADADA;
	-moz-box-shadow: 1px 1px 4px #DADADA;
	-webkit-box-shadow: 1px 1px 4px #DADADA;
	border-radius: 3px;
	-webkit-border-radius: 3px;
	-moz-border-radius: 3px;
}
.form-style-2 input[type=submit]:hover,
.form-style-2 input[type=button]:hover{
	background: #EA7B00;
	color: #fff;
}
</style>

<body>
<form>
<div class="form-style-2">
<div class="form-style-2-heading">Checkout files in the QTechNG repository</div>
<form>
<label for="qpattern"><span>Qpatterns <span class="required">*</span></span><input autofocus placeholder="e.g. /core/python3/*.py" type="text" class="input-field" name="qpath" id="qpattern" value="" /></label>
<label for="version"><span>Version <span class="required">*</span></span><input type="text" class="input-field" name="version" id="version" value="0.00" /></label>
<label><span> </span><input type="submit" value="Submit" onclick="golangfunc()" /></label>
</form>




		<pre id="response"></pre>
		</body>
	</html>
	`), "", 480, 320)
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
