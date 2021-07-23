package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	qtranslate "cloud.google.com/go/translate/apiv3"
	"github.com/spf13/cobra"
	qtranslatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"
)

var textTranslateCmd = &cobra.Command{
	Use:   "translate",
	Short: "translate text",
	Long: `Command which translates text.
The text is specified as the first and only argument.
If there are no arguments, the text to be translated, is retrieved from stdin.

The '--lgsource' flag specifies the source language, according to BCP-47 (See: https://www.rfc-editor.org/info/bcp47). 
Use the '--lgsource' flag. Default value is "nl-NL"

The '--lgtarget' flag specifies the target language(s) according to BCP-47. Default value is "en-GB,fr-FR"

If the '--isjson' flag is present, the argument is interpreted as a file with 
a JSON array. Every element of the array is translated.`,
	Args: cobra.MaximumNArgs(1),
	Example: `qtechng text translate "Opgelet ! Er staan cijfers in de auteursnaam en dit is GEEN authority code"
qtechng text translate translateme.json --isjson`,

	RunE:   textTranslate,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
	},
}

var Flgsource = ""
var Flgtarget = ""
var Fisjson = false

func init() {
	textTranslateCmd.Flags().StringVar(&Flgsource, "lgsource", "", "Brontaal")
	textTranslateCmd.Flags().StringVar(&Flgtarget, "lgtarget", "", "Bestemmingstaal")
	textTranslateCmd.Flags().BoolVar(&Fisjson, "isjson", false, "is het argument een JSON bestand")
	textCmd.AddCommand(textTranslateCmd)
}

func textTranslate(cmd *cobra.Command, args []string) error {

	text := ""
	if len(args) == 0 {
		btext, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		text = strings.TrimSpace(string(btext))
	} else {
		text = args[0]
	}

	if text == "" {
		return nil
	}

	if Flgsource == "" {
		Flgsource = "nl-NL"
	}
	if Flgtarget == "" {
		Flgtarget = "en-GB,fr-FR"
	}

	lgtargets := strings.SplitN(Flgtarget, ",", -1)

	results := make([]map[string]string, 0)
	errs := make([]error, 0)

	texts := make([]string, 0)
	if Fisjson {
		data, err := qfs.Fetch(qutil.AbsPath(text, Fcwd))
		if err != nil {
			Fmsg = qreport.Report(results, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		}
		err = json.Unmarshal(data, &texts)
		if err != nil {
			Fmsg = qreport.Report(results, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		}
	} else {
		texts = append(texts, text)
	}

	for _, text := range texts {
		one := make(map[string]string)
		results = append(results, one)
		one[Flgsource] = text
		for _, lg := range lgtargets {
			lg = strings.TrimSpace(lg)
			if lg == "" {
				continue
			}
			r, e := translate(text, Flgsource, lg)
			if e != nil {
				errs = append(errs, e)
				continue
			}
			one[lg] = string(r)
		}
	}
	if len(errs) == 0 {
		errs = nil
	}
	if len(results) == 0 {
		results = nil
	}
	Fmsg = qreport.Report(results, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil

}

func translate(text string, from string, to string) (tr string, err error) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", qregistry.Registry["google-translate-privatekey"])
	ctx := context.Background()
	c, err := qtranslate.NewTranslationClient(ctx)
	if err != nil {
		return
	}
	defer c.Close()
	request := qtranslatepb.TranslateTextRequest{
		Contents:           []string{text},
		MimeType:           "text/plain",
		SourceLanguageCode: from,
		TargetLanguageCode: to,
		Parent:             "projects/translation-of-lgcodes",
	}
	response, err := c.TranslateText(ctx, &request)
	if err != nil {
		return
	}
	translations := response.Translations
	if len(translations) == 0 {
		return
	}
	trl := translations[0]
	tr = trl.TranslatedText

	return

}
