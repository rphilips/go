package text

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	qregistry "brocade.be/base/registry"
	qtranslate "cloud.google.com/go/translate/apiv3"
	qtranslatepb "google.golang.org/genproto/googleapis/cloud/translate/v3"
)

func GoogleTranslate(text string, from string, to string) (tr string, err error) {
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

func DeepLTranslate(text string, from string, to string) (tr string, err error) {
	to = strings.ToUpper(to)
	if strings.ContainsRune(to, '-') && !strings.HasPrefix(to, "EN-") {
		to = strings.SplitN(to, "-", -1)[0]
	}
	from = strings.ToUpper(from)
	if strings.ContainsRune(from, '-') && !strings.HasPrefix(from, "EN-") {
		from = strings.SplitN(from, "-", -1)[0]
	}
	auth_key := qregistry.Registry["deepl-translate-privatekey"]
	resp, err := http.PostForm("https://api-free.deepl.com/v2/translate",
		url.Values{
			"auth_key":            []string{auth_key},
			"text":                []string{text},
			"target_lang":         []string{to},
			"source_lang":         []string{from},
			"preserve_formatting": []string{"1"},
		})
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if len(body) == 0 {
		return "", errors.New("no translation found: empty response")
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(body, &m)

	if err != nil {
		return
	}
	trs, ok := m["translations"]
	if !ok || trs == nil {
		return "", errors.New("no translation found [1]")
	}
	switch v := trs.(type) {
	case []interface{}:
		if len(v) == 0 || v[0] == nil {
			return "", errors.New("no translation found [2]")
		}
		switch w := v[0].(type) {
		case map[string]interface{}:
			tr := w["text"]
			str := tr.(string)
			if str == "" {
				return "", errors.New("no translation found [3]")
			}
			return str, nil
		default:
			return "", errors.New("no translation found [4]")
		}
	default:
		return "", errors.New("no translation found [5]")
	}

}
