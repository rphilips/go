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

var textDetectCmd = &cobra.Command{
	Use:   "detect",
	Short: "detect language",
	Long: `Command which detects the language of a text.
The text is specified as the first and only argument.
If there are no arguments, the text to be examined, is retrieved from stdin.

If the *--isjson flag* is present, the argument is interpreted as a file with 
a JSON array. Every element of the array is examined.`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng text detect "Goede morgen"`,

	RunE:   textDetect,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
	},
}

func init() {
	textDetectCmd.Flags().BoolVar(&Fisjson, "isjson", false, "is het argument een JSON bestand")
	textCmd.AddCommand(textDetectCmd)
}

type lgdetect struct {
	Text string             `json:"text"`      // Text
	Lgs  map[string]float32 `json:"languages"` // probabilities
}

func textDetect(cmd *cobra.Command, args []string) error {

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

	results := make([]lgdetect, 0)
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

		lgs, err := detect(text)
		if err == nil {
			results = append(results, lgdetect{text, lgs})
		}
		if err != nil {
			errs = append(errs, err)
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

func detect(text string) (probs map[string]float32, err error) {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", qregistry.Registry["google-translate-privatekey"])
	ctx := context.Background()
	c, err := qtranslate.NewTranslationClient(ctx)
	if err != nil {
		return
	}
	defer c.Close()
	request := qtranslatepb.DetectLanguageRequest{
		MimeType: "text/plain",
		Source:   &qtranslatepb.DetectLanguageRequest_Content{Content: text},
		Parent:   "projects/translation-of-lgcodes",
	}
	response, err := c.DetectLanguage(ctx, &request)
	if err != nil {
		return
	}
	languages := response.Languages
	if len(languages) == 0 {
		return
	}
	probs = make(map[string]float32)

	for _, lg := range languages {
		code := lg.LanguageCode
		prob := lg.Confidence
		if code != "" {
			probs[code] = prob
		}
	}

	return

}
