package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qtext "brocade.be/qtechng/lib/text"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var textTranslateCmd = &cobra.Command{
	Use:   "translate",
	Short: "Translate text",
	Long: `This command translates text.
The text is specified as the first and only argument.
If there are no arguments, the text to be translated, is retrieved from stdin.

The '--lgsource' flag specifies the source language, according to BCP-47
(See: https://www.rfc-editor.org/info/bcp47).

Use the '--lgsource' flag. Default value is "nl-NL"

The '--lgtarget' flag specifies the target language(s) according to BCP-47.
Default value is "en-GB,fr-FR"

If the '--isfile' flag is present, the argument is interpreted as a file with
a JSON array. Every element of the array is translated.`,
	Args: cobra.MaximumNArgs(1),
	Example: `qtechng text translate "Opgelet ! Er staan cijfers in de auteursnaam en dit is GEEN authority code"
qtechng text translate translateme.json --isfile`,
	RunE:   textTranslate,
	PreRun: textTranslateLocal,
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
	},
}

var Flgsource = ""
var Flgtarget = ""
var Fisfile = false

func init() {
	textTranslateCmd.Flags().StringVar(&Flgsource, "lgsource", "", "Brontaal")
	textTranslateCmd.Flags().StringVar(&Flgtarget, "lgtarget", "", "Bestemmingstaal")
	textTranslateCmd.Flags().BoolVar(&Fisfile, "isfile", false, "is het argument een JSON bestand")
	textCmd.AddCommand(textTranslateCmd)
}

func textTranslateLocal(cmd *cobra.Command, args []string) {
	if strings.ContainsRune(qregistry.Registry["qtechng-type"], 'B') {
		return
	}
	if len(args) == 0 {
		btext, err := io.ReadAll(os.Stdin)
		if err != nil {
			return
		}
		text := strings.TrimSpace(string(btext))
		if text != "" {
			args = append(args, text)
		}
	}
	if len(args) == 0 {
		return
	}
	omit := args[0]
	if Fisfile {
		data, err := qfs.Fetch(qutil.AbsPath(args[0], Fcwd))
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return
		}
		err = json.Unmarshal(data, &args)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return
		}
	}
	if len(args) == 0 {
		return
	}
	argums := make([]string, 0)
	argums = append(argums, os.Args[0], os.Args[1], os.Args[2])
	argums = append(argums, args...)
	for _, a := range os.Args[3:] {
		if strings.HasPrefix(a, "--isfile=") {
			continue
		}
		if a == omit {
			continue
		}
		argums = append(argums, a)
	}
	os.Args = argums
	preSSH(cmd, nil)
}

func textTranslate(cmd *cobra.Command, args []string) error {

	services := qregistry.Registry["qtechng-translation-services"]
	if services == "" {
		Fmsg = qreport.Report(nil, errors.New("no translation services defined"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	trsystems := strings.SplitN(services, ",", -1)

	text := ""
	if len(args) == 0 {
		btext, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		text = strings.TrimSpace(string(btext))
		args = append(args, text)
	}
	if len(args) == 0 {
		return nil
	}

	if Flgsource == "" {
		Flgsource = "nl-NL"
	}
	if Flgtarget == "" {
		Flgtarget = "en-GB,fr-FR"
	}

	lgtargets := strings.SplitN(Flgtarget, ",", -1)

	if Fisfile {
		data, err := qfs.Fetch(qutil.AbsPath(args[0], Fcwd))
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
		err = json.Unmarshal(data, &args)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
	}
	if len(args) == 0 {
		return nil
	}

	type mission struct {
		From        string `json:"lgsource"`
		To          string `json:"lgtarget"`
		Text        string `json:"text"`
		System      string `json:"system"`
		Translation string `json:"translation"`
		Error       string `json:"error"`
	}

	missions := make([]mission, 0)

	for _, lg := range lgtargets {
		lg = strings.TrimSpace(lg)
		if lg == "" {
			continue
		}
		for _, text := range args {
			if text == "" {
				continue
			}
			for _, system := range trsystems {

				missions = append(missions, mission{
					From:   Flgsource,
					To:     lg,
					Text:   text,
					System: system,
				})
			}
		}
	}

	fn := func(n int) (interface{}, error) {
		m := missions[n]
		system := m.System
		tr := ""
		var err error
		switch system {
		case "google":
			tr, err = qtext.GoogleTranslate(m.Text, m.From, m.To)
		case "deepl":
			tr, err = qtext.DeepLTranslate(m.Text, m.From, m.To)
		}
		m.Translation = tr
		if err != nil {
			m.Error = err.Error()
		}
		return m, nil
	}

	results, _ := qparallel.NMap(len(missions), -1, fn)

	Fmsg = qreport.Report(results, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")

	return nil

}
