package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
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
var Fdir string

func init() {
	newCmd.PersistentFlags().BoolVar(&Fstderr, "stderr", false, "Show stderr")
	newCmd.PersistentFlags().StringVar(&Fdir, "dir", "", "Directory")
	rootCmd.AddCommand(newCmd)
}

func newedition(cmd *cobra.Command, args []string) error {
	if !Fstderr {
		_, err := ptools.Launch([]string{"gopblad", "new", "--stderr", "--dir=" + Fdir}, nil, "", false)
		return err
	}
	install(cmd, args)

	// HTML
	t, err := template.ParseFS(guifs, "templates/new.html")
	if err != nil {
		return err
	}
	mold := loadVars(&guiFiller)
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
		})

		ui.Run()
		closed := map[string]string{"action": "close"}
		guilistener <- closed
	}()

	mm := <-guilistener
	if mm["action"] != "new" {
		return nil
	}

	m, err := pmanuscript.New(mm, mold)

	if err == nil {
		if Fdir == "" {
			Fdir = pfs.FName("workspace")
		}
		weekpb := filepath.Join(Fdir, "week.pb")
		source := strings.NewReader(m.String())
		m, err := pmanuscript.Parse(source, false, "")
		if err != nil {
			bfs.Store("/home/rphilips/Desktop/log.js", err.Error(), "process")
			return err
		}
		bfs.Store(weekpb, m.String(), "process")
	}

	return err
}

func loadVars(guiFiller *GuiFiller) (mold *pmanuscript.Manuscript) {
	mold = pmanuscript.Previous()

	id := "?"
	period := "?"
	mailed := "?"
	if mold != nil {
		id = mold.ID()
		period = fmt.Sprintf("%s - %s", ptools.StringDate(mold.Bdate, "I"), ptools.StringDate(mold.Edate, "I"))
		mailed = ptools.StringDate(mold.Edate, "I")
	}
	_, year, week, bdate, edate := pmanuscript.Next(mold)

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
	return
}

func handleNew(action string, ui webview.WebView, guiFiller *GuiFiller) map[string]string {
	mm := make(map[string]string)
	json.Unmarshal([]byte(action), &mm)
	return mm
}
