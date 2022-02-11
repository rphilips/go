package html

import (
	"bytes"
	"text/template"
)

// Make HTML result page
func Result(keys interface{}) string {
	result := `<!DOCTYPE html>
	<html>

	<head>
		<meta charset="UTF-8" />
	</head>

	<body>
		<div><input type="button" onclick="location.href='{{ .BaseURL}}';" value="back" /></div>
		<br>
		<div><tt>{{ .Qresponse}}</tt></div>
	</body>

	</html>`

	ut, err := template.New("result").Parse(result)
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
