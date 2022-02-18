package html

import (
	"bytes"
	"text/template"

	css "brocade.be/qui/lib/css"
)

// Make HTML result page
func Delete(keys interface{}) string {
	delete := `<!DOCTYPE html>
	<html>

	<head>
		<meta charset="UTF-8" />
		<style>` + css.CSS + `</style>
	</head>

	<body>
		<div><input type="button" onclick="location.href='{{ .BaseURL}}';" value="back" /></div>
		<br>
		<form method="POST" action="/result" autocomplete="off">
			<input name="cmd" id="cmd" type="hidden" value="" />

			<fieldset><legend><b>Delete</b></legend>
				<input type="submit" value="delete" onclick="document.getElementById('cmd').value='delete'" />
			</fieldset>

		</form>
	</body>

	</html>`

	ut, err := template.New("result").Parse(delete)
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