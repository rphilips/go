package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"path"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
	"github.com/zserge/lorca"
)

//go:embed templates
var guifs embed.FS

// Finplace replace the file contents

var guiCmd = &cobra.Command{
	Use:     "gui",
	Short:   "GUI functions",
	Long:    `All kinds of GUI functions`,
	Args:    cobra.NoArgs,
	RunE:    guiMenu,
	Example: "qtechng gui",
}

type GuiFiller struct {
	CSS   template.CSS
	JS    template.JS
	Head  template.HTML
	Menu  bool
	Title string
	Vars  map[string]string
	VarsH map[string]bool
}

var guiFiller GuiFiller

//Fmenu with menu
var Fmenu string

func init() {
	rootCmd.AddCommand(guiCmd)
	guiCmd.PersistentFlags().StringVar(&Fmenu, "menu", "", "Menu identifier")
	guiFiller.Title = "QtechNG"

	b, _ := guifs.ReadFile("templates/gui.css")
	guiFiller.CSS = template.CSS(b)
	b, _ = guifs.ReadFile("templates/gui.js")
	guiFiller.JS = template.JS(b)
	head := `<title>{{ .Title }}</title><style type="text/css">{{ .CSS }}</style><script type="text/javascript">{{ .JS }}</script>`
	t, _ := template.New("foo").Parse(head)
	buf := new(bytes.Buffer)
	_ = t.Execute(buf, guiFiller)
	guiFiller.Head = template.HTML(buf.Bytes())
}

func guiMenu(cmd *cobra.Command, args []string) error {
	standalone := Fmenu != "menu" && Fmenu != ""
	ui, err := lorca.New("", "", 520, 640)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	menulistener := make(chan string)
	defer close(menulistener)
	go func() {
		prevMenu := ""
		for {
			menuitem := ""
			if Fmenu == "" {
				Fmenu = "menu"
			}
			if prevMenu != Fmenu {
				loadVars(Fmenu, &guiFiller)
				t, err := template.ParseFS(guifs, "templates/"+Fmenu+".html")
				if err != nil {
					Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
					return
				}
				buf := new(bytes.Buffer)
				err = t.Execute(buf, guiFiller)

				if err != nil {
					Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
					return
				}
				ui.Load("data:text/html," + url.PathEscape(buf.String()))

				switch Fmenu {
				case "menu":
					ui.Bind("golangfunc", func(indicator string) {
						menuitem = indicator
						if menuitem != "" {
							menulistener <- menuitem
						}
					})

				case "search":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}

						handleSearch(ui, &guiFiller)
						menulistener <- "search"
					})

				}
			}
			prevMenu = Fmenu
			menuitem = <-menulistener
			if menuitem == "stop" {
				switch {
				case Fmenu == "menu":
					ui.Eval(`window.close()`)
					ui.Close()
					return
				case standalone:
					ui.Eval(`window.close()`)
					ui.Close()
					return
				default:
					Fmenu = "menu"
				}
			} else {

				Fmenu = menuitem
			}
		}
	}()
	<-ui.Done()
	return nil
}

func loadVars(menu string, guiFiller *GuiFiller) {
	vars := make(map[string]string)
	varsH := make(map[string]bool)
	fname := path.Join(qregistry.Registry["scratch-dir"], "menu-"+menu+".json")
	data, err := qfs.Fetch(fname)
	if err == nil {
		json.Unmarshal(data, &vars)
	}
	for _, k := range []string{"perline", "yaml"} {
		_, ok := vars[k]
		if !ok {
			vars[k] = "1"
		}
		if vars[k] == "1" {
			varsH[k] = true
		}
	}
	for _, k := range []string{"tolower", "regexp", "checkout", "clear"} {
		_, ok := vars[k]
		if !ok {
			vars[k] = "0"
		}
		if vars[k] == "1" {
			varsH[k] = true
		}
	}

	_, ok := vars["version"]
	if !ok {
		vars["version"] = "0.00"
	}

	_, ok = vars["editlist"]
	if !ok {
		vars["editlist"] = menu
	}

	_, ok = vars["jsonpath"]
	if !ok {
		vars["jsonpath"] = "$..qpath"
	}
	guiFiller.Vars = vars
	guiFiller.VarsH = varsH

}

func storeVars(menu string, guiFiller GuiFiller) {
	fname := path.Join(qregistry.Registry["scratch-dir"], "menu-"+menu+".json")
	b, err := json.Marshal(guiFiller.Vars)
	if err != nil {
		return
	}
	qfs.Store(fname, b, "")
}

func handleSearch(ui lorca.UI, guiFiller *GuiFiller) {
	f := make(map[string]string)
	for _, key := range []string{"qpattern", "version", "needle", "perline", "tolower", "regexp", "checkout", "yaml", "jsonpath", "editlist", "clear"} {
		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
		f[key] = value
	}
	guiFiller.Vars = f
	storeVars("search", *guiFiller)

	if f["qpattern"] == "" || f["version"] == "" {
		return
	}
	argums := []string{
		"source",
	}
	if f["checkout"] == "1" {
		argums = append(argums, "co", "--auto")
		if f["clear"] == "1" {
			argums = append(argums, "--clear")
		}
	} else {
		argums = append(argums, "list")
	}
	argums = append(argums, "--version="+f["version"])
	argums = append(argums, "--qpattern="+f["qpattern"])
	argums = append(argums, "--needle="+f["needle"])
	argums = append(argums, "--list="+f["editlist"])
	if f["perline"] == "1" {
		argums = append(argums, "--perline")
	}
	if f["tolower"] == "1" {
		argums = append(argums, "--tolower")
	}
	if f["regexp"] == "1" {
		argums = append(argums, "--regexp")
	}
	ui.Eval(`document.getElementById("busy").innerHTML = "Busy ..."`)
	ui.Eval(`document.getElementById("busy").style="display:block;")`)
	sout, serr, err := qutil.QtechNG(argums, f["jsonpath"], f["yaml"] == "1")
	ui.Eval(`document.getElementById("busy").innerHTML = ""`)
	fmt.Println(sout)
	fmt.Println(serr)
	if err != nil {
		serr += "\n\nError:" + err.Error()
	}
	bx, _ := json.Marshal(sout)
	sx := string(bx)
	st := ""
	if len(sx) > 2 {
		st = sx[:2]
	}
	k := strings.IndexAny(st, "[{")
	if k == 1 {
		ui.Eval(`document.getElementById("jsondisplay").innerHTML = syntaxHighlight(` + sx + `)`)
		ui.Eval(`document.getElementById("yamldisplay").innerHTML = ""`)
	} else {
		ui.Eval(`document.getElementById("yamldisplay").innerHTML = ` + sx)
		ui.Eval(`document.getElementById("jsondisplay").innerHTML = ""`)
	}
}
