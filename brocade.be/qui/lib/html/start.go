package html

import (
	"bytes"
	"text/template"

	css "brocade.be/qui/lib/css"
	"brocade.be/qui/lib/js"
)

// Make HTML start page
func Start(keys interface{}) string {
	start := `<!DOCTYPE html>
	<html>

	<head>
		<meta charset="UTF-8" />
		<style>` + css.CSS + `</style>
	</head>

	<body>
		<table>
		<tr><td><div style=overflow-y:auto;height:300px;width:200px">{{ range .Qpaths }}{{ . }}{{ end }}</div></td>
		<td style=vertical-align:top>
			<h1>QtechNG</h1>
			{{ .Name}}
			 <form method="POST" action="/result" autocomplete="off">
				<input name="cmd" id="cmd" type="hidden" value="" />

				<fieldset><legend><b>Info</b></legend>
				<input type="submit" value="about" onclick="document.getElementById('cmd').value='about'" />
				<input type="submit" value="quit" onclick="document.getElementById('cmd').value='quit'" />
				</fieldset>

				<fieldset><legend><b>Files</b></legend>
				<div class="autocomplete">
					<input id="file" type="text" name="file" size="100">
				</div>
				<p>

				<input type="submit" value="tell" onclick="document.getElementById('cmd').value='tell'" />
				<input type="submit" value="touch" onclick="document.getElementById('cmd').value='touch'" />
				<input type="submit" value="open" onclick="document.getElementById('cmd').value='open'" />
				<input type="submit" value="checkin" onclick="document.getElementById('cmd').value='checkin'" />
				<input type="submit" value="checkout" onclick="document.getElementById('cmd').value='checkout'" />
				</fieldset>

				<fieldset><legend><b>Compare</b></legend>
				<input type="submit" value="previous" onclick="document.getElementById('cmd').value='previous'" />
				<input type="submit" value="git" onclick="document.getElementById('cmd').value='git'" />
				</fieldset>

			</form>

			<fieldset><legend><b>Repository</b></legend>
				<input id="path" type="text" name="path" size="100">

				<form method="POST" action="/rename" autocomplete="off">
					<input name="cmd" id="cmd" type="hidden" value="" />
					<input type="submit" value="rename" onclick="document.getElementById('cmd').value='rename'" />
				</form> <br>

				<form method="POST" action="/delete" autocomplete="off">
					<input name="cmd" id="cmd" type="hidden" value="" />
					<input type="submit" value="delete" onclick="document.getElementById('cmd').value='delete'" />
				</form>

			</fieldset>

			<form method="POST" action="/result" autocomplete="off">
				<input name="cmd" id="cmd" type="hidden" value="" />

				<fieldset><legend><b>System</b></legend>
				<input type="submit" value="registry" onclick="document.getElementById('cmd').value='registry'" />
				<input type="submit" value="setup" onclick="document.getElementById('cmd').value='setup'" />
				<input type="submit" value="commands" onclick="document.getElementById('cmd').value='commands'" />
				<input type="submit" value="checkout all" onclick="document.getElementById('cmd').value='all'" />
				</fieldset>

				<fieldset><legend><b>Links</b></legend>
				<a href="https://dev.anet.be/brocade">presto</a><br>
				<a href="https://anet.be/brocade">moto</a><br>
				<a href="https://anetbrocade.slack.com/">slack</a><br>
				<a href="https://anet.be/probes/probes.phtml">probes</a>
				</fieldset>

			</form>

		</td></tr>
		</table>
			<script>` + js.JS + `
				var qfiles = [{{ range .Qfiles }}{{ print "\"" }}{{ . }}{{ print "\"" }}{{ print "," }}{{ end }}];
				autocomplete(document.getElementById("file"), qfiles);
			</script>
	</body>

	</html>`

	ut, err := template.New("start").Parse(start)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = ut.Execute(&tpl, keys)
	if err != nil {
		panic(err)
	}

	return tpl.String()
}
