package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	btime "brocade.be/base/time"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
	ptools "brocade.be/pbladng/lib/tools"
	"github.com/spf13/cobra"
	"github.com/webview/webview"
)

//go:embed templates
var guifs embed.FS

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Make a new edition",
	Long:  "Make a new edition - based on the previous one - of pblad",

	Args:    cobra.NoArgs,
	Example: `pblad new`,
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
		_, err := ptools.Launch([]string{"gopblad", "new", "--stderr", "--cwd=" + Fcwd}, nil, "", false, false)
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

	weekpb := filepath.Join(Fcwd, "parochieblad.ed")
	doc, err := pstructure.New(mm, mold)
	doc.Dir = Fcwd
	if err == nil {
		source := strings.NewReader(doc.String())
		err := doc.Load(source)
		if err != nil {
			ptools.Log(err)
			fmt.Println(err)
			return err
		}
		bfs.Store(weekpb, doc.String(), "process")
		pedit := pregistry.Registry["editor-exe"].([]any)
		edit := make([]string, 0)
		for _, piece := range pedit {
			p := piece.(string)
			p = strings.ReplaceAll(p, "{fname}", weekpb)
			edit = append(edit, p)
		}
		vcmd := exec.Command(edit[0], edit[1:]...)
		err = vcmd.Start()
		if err != nil {
			return err
		}
		fmt.Println(weekpb)

	}

	return err
}

func loadVars(guiFiller *GuiFiller) (mold *pstructure.Document) {
	_, mold, _ = pstructure.FindBefore("")
	id := "?"
	period := "?"
	mailed := "?"
	year := "?"
	week := "?"
	bdate := "?"
	edate := "?"
	if mold != nil {
		id = mold.ID()
		period = fmt.Sprintf("%s - %s", btime.StringDate(mold.Bdate, "I"), btime.StringDate(mold.Edate, "I"))
		mailed = btime.StringDate(mold.Mailed, "I")
		_, syear, sweek, tbdate, tedate := mold.Next()
		if syear != "" {
			year = syear
		}
		if sweek != "" {
			week = sweek
		}
		if tbdate != nil {
			bdate = btime.StringDate(tbdate, "I")
		}
		if tedate != nil {
			edate = btime.StringDate(tedate, "I")
		}
	}

	minbdate := bdate
	maxbdate := bdate
	minyear := "2005"
	now := time.Now()
	maxyear := strconv.Itoa(now.Year() + 1)
	edatedisplay := ""
	if edate != "" {
		t := btime.DetectDate(edate)
		edatedisplay = btime.StringDate(t, "D")
	}
	bdatedisplay := ""
	if bdate != "" {
		t := btime.DetectDate(bdate)
		bdatedisplay = btime.StringDate(t, "D")
	}

	vars := map[string]string{
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
