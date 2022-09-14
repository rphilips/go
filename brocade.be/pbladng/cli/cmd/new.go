package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"

	pmanuscript "brocade.be/pbladng/lib/manuscript"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
	"github.com/spf13/cobra"
	"github.com/webview/webview"

	// "github.com/zserge/lorca"
	bfs "brocade.be/base/fs"
)

//go:embed templates
var guifs embed.FS

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "new `gopblad`",
	Long:  "new `gopblad`",

	Args:    cobra.NoArgs,
	Example: `gopblad new`,
	RunE:    newedition,
}

type GuiFiller struct {
	Vars map[string]string
}

var guiFiller GuiFiller
var Fstderr bool

func init() {
	newCmd.PersistentFlags().BoolVar(&Fstderr, "stderr", false, "Show stderr")
	rootCmd.AddCommand(newCmd)
}

func newedition(cmd *cobra.Command, args []string) error {
	if !Fstderr {
		_, err := ptools.Launch([]string{"gopblad", "new", "--stderr"}, nil, "", false)
		return err
	}
	install(cmd, args)

	// HTML
	t, err := template.ParseFS(guifs, "templates/new.html")
	if err != nil {
		return err
	}
	loadVars(&guiFiller)
	buf := new(bytes.Buffer)
	err = t.Execute(buf, guiFiller)

	if err != nil {
		return err
	}

	// interface
	guilistener := make(chan map[string]string)
	defer close(guilistener)
	os.Stderr, _ = os.Open(os.DevNull)
	go func() {
		os.Stderr, _ = os.Open(os.DevNull)
		width := 520
		height := 780

		ui := webview.New(false)
		defer ui.Destroy()

		ui.SetSize(width, height, webview.HintNone)
		ui.SetHtml(buf.String())
		ui.Bind("register", func(action string) {
			bfs.Store("/home/rphilips/Desktop/log.txt", action, "process")
			mm := handleNew(action, ui, &guiFiller)
			guilistener <- mm
			ui.Eval("window.close()")
			ui.Terminate()
			return
		})

		ui.Run()
		closed := map[string]string{"action": "close"}
		guilistener <- closed
	}()

	mm := <-guilistener
	bfs.Store("/home/rphilips/Desktop/log.txt", fmt.Sprintf("%v", mm), "process")
	return nil
}

func loadVars(guiFiller *GuiFiller) {
	id, period, mailed := pmanuscript.Previous()
	year, week, bdate, edate := pmanuscript.Next()
	tbdate, _, _ := ptools.NewDate(bdate)
	tedate, _, _ := ptools.NewDate(edate)
	vars := map[string]string{
		"pcode":        pregistry.Registry["pcode"].(string),
		"id":           id,
		"periode":      period,
		"mailed":       mailed,
		"year":         year,
		"week":         week,
		"bdate":        bdate,
		"edate":        edate,
		"minyear":      "2000",
		"maxyear":      "2050",
		"minweek":      "1",
		"maxweek":      "53",
		"minbdate":     bdate,
		"maxbdate":     "2010-06-30",
		"edatedisplay": ptools.StringDate(tedate, "D"),
		"bdatedisplay": ptools.StringDate(tbdate, "D"),
	}
	guiFiller.Vars = vars

}

func handleNew(action string, ui webview.WebView, guiFiller *GuiFiller) map[string]string {
	mm := make(map[string]string)
	json.Unmarshal([]byte(action), &mm)
	return mm
}
