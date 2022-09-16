package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"time"

	pmanuscript "brocade.be/pbladng/lib/manuscript"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
	"github.com/spf13/cobra"
	"github.com/webview/webview"
	// "github.com/zserge/lorca"
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
	fmt.Println(mm)
	return nil
}

func loadVars(guiFiller *GuiFiller) {
	m := pmanuscript.Previous()

	id := "?"
	period := "?"
	mailed := "?"
	if m != nil {
		id = m.ID()
		period = fmt.Sprintf("%s - %s", ptools.StringDate(m.Bdate, "I"), ptools.StringDate(m.Edate, "I"))
		mailed = ptools.StringDate(m.Edate, "I")
	}
	_, year, week, bdate, edate := pmanuscript.Next(m)

	minbdate := bdate
	maxbdate := bdate
	minyear := "2005"
	now := time.Now()
	maxyear := strconv.Itoa(now.Year() + 1)
	edatedisplay := ""
	if edate != "" {
		t, _, _ := ptools.NewDate(edate)
		edatedisplay = ptools.StringDate(t, "D")
	}
	bdatedisplay := ""
	if bdate != "" {
		t, _, _ := ptools.NewDate(bdate)
		bdatedisplay = ptools.StringDate(t, "D")
	}

	vars := map[string]string{
		"pcode":        pregistry.Registry["pcode"].(string),
		"id":           id,
		"periode":      period,
		"mailed":       mailed,
		"year":         year,
		"week":         week,
		"bdate":        bdate,
		"edate":        edate,
		"minyear":      minyear,
		"maxyear":      maxyear,
		"minweek":      "1",
		"maxweek":      "53",
		"minbdate":     minbdate,
		"maxbdate":     maxbdate,
		"edatedisplay": edatedisplay,
		"bdatedisplay": bdatedisplay,
	}
	guiFiller.Vars = vars

}

func handleNew(action string, ui webview.WebView, guiFiller *GuiFiller) map[string]string {
	mm := make(map[string]string)
	json.Unmarshal([]byte(action), &mm)
	return mm
}
