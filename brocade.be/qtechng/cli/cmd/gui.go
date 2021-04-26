package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
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
	Args:    cobra.MinimumNArgs(0),
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
	VarsS map[string]template.HTML
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
	if len(args) == 1 && (args[0] == Fcwd || qfs.SameFile(args[0], Fcwd)) {
		args = args[1:]
	}
	standalone := Fmenu != "menu" && Fmenu != ""
	width := 520
	height := 780

	switch Fmenu {
	case "property":
		width = 680
	}

	ui, err := lorca.New("", "", width, height)
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
				// things todo before menu is loaded
				switch Fmenu {
				case "checkin":
					sout := handleCheckin(Fcwd, args)
					bx, _ := json.Marshal(sout)
					sx := string(bx)
					st := ""
					if len(sx) > 2 {
						st = sx[:2]
					}
					k := strings.IndexAny(st, "[{")
					if k == 1 {
						guiFiller.Vars["jsondisplay"] = sx
						guiFiller.Vars["yamldisplay"] = ""
					} else {
						guiFiller.Vars["jsondisplay"] = ""
						guiFiller.Vars["yamldisplay"] = sx
					}

				case "property":
					if len(args) > 0 {
						fname := qutil.AbsPath(args[0], Fcwd)
						guiFiller.Vars["fname"] = fname
						argums := []string{
							"file",
							"list",
							fname,
						}
						out, _, _ := qutil.QtechNG(argums, "$..DATA", false, Fcwd)
						guiFiller.Vars["properties"] = string(out)
					}
				case "new":
					mydir := Fcwd
					argums := []string{
						"dir",
						"tell",
						"--cwd=" + mydir,
					}
					out, _, _ := qutil.QtechNG(argums, "$..DATA", false, Fcwd)
					m := make(map[string]string)
					json.Unmarshal([]byte(out), &m)
					qdir := m["qdir"]
					version := m["version"]
					if version == "" {
						version = "0.00"
					}
					guiFiller.Vars["qdir"] = qdir
					guiFiller.Vars["version"] = version
					guiFiller.VarsS = make(map[string]template.HTML)
					guiFiller.VarsH["nofiles"] = len(args) == 0
					guiFiller.VarsS["select"] = template.HTML(hintoptions(Fcwd))

				}

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
				case "property":
					sx := guiFiller.Vars["properties"]
					if sx != "" {
						bx, _ := json.Marshal(sx)
						sx := string(bx)
						ui.Eval(`document.getElementById("jsondisplay").innerHTML = syntaxHighlight(` + sx + `)`)
					}
				case "checkin":
					sx := guiFiller.Vars["jsondisplay"]
					if sx != "" {
						ui.Eval(`document.getElementById("jsondisplay").innerHTML = syntaxHighlight(` + sx + `)`)
						ui.Eval(`document.getElementById("yamldisplay").innerHTML = ""`)
					}
					sx = guiFiller.Vars["yamldisplay"]
					if sx != "" {
						ui.Eval(`document.getElementById("yamldisplay").innerHTML = ` + sx)
						ui.Eval(`document.getElementById("jsondisplay").innerHTML = ""`)
					}
				}

				switch Fmenu {

				case "menu":
					ui.Bind("golangfunc", func(indicator string) {
						menuitem = indicator
						if menuitem != "" {
							menulistener <- menuitem
						}
					})

				case "checkin":
					ui.Bind("golangfunc", func(indicator string) {
						menulistener <- "stop"
					})

				case "search":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						handleSearch(ui, &guiFiller)
						menulistener <- handleSearch(ui, &guiFiller)
					})
				case "checkout":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						handleSearch(ui, &guiFiller)
						menulistener <- handleCheckout(ui, &guiFiller, args)
					})
				case "property":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						menulistener <- handleProperty(ui, &guiFiller)
					})

				case "new":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						menulistener <- handleNew(ui, &guiFiller, Fcwd, args)
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
	qfs.Store(fname, b, "qtech")
}

func handleCheckin(cwd string, args []string) string {
	argums := []string{
		"file",
		"ci",
		"--cwd=" + cwd,
		"--recurse",
	}
	if len(args) != 0 {
		argums = append(argums, args...)
	}
	sout, _, _ := qutil.QtechNG(argums, "$..file", true, cwd)
	return sout
}

func handleNew(ui lorca.UI, guiFiller *GuiFiller, cwd string, args []string) string {
	f := make(map[string]string)
	keys := []string{"qdir", "version"}
	if len(args) == 0 {
		keys = append(keys, "name", "hint")
	}
	for _, key := range keys {
		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
		f[key] = value
	}
	if f["version"] == "" || f["qdir"] == "" {
		return "stop"
	}
	argums := make([]string, 0)
	if len(args) == 0 {
		f := make(map[string]string)
		for _, key := range []string{"name", "qdir", "version", "hint"} {
			value := ui.Eval(`document.getElementById('` + key + `').value`).String()
			f[key] = value
		}
		if f["name"] == "" || f["version"] == "" || f["qdir"] == "" {
			return "new"
		}
		argums = []string{
			"file",
			"new",
			"--version=" + f["version"],
			"--qdir=" + f["qdir"],
			"--create",
			f["name"],
			"--cwd=" + cwd,
		}
		if f["hint"] != "" {
			argums = append(argums, "--hint="+f["hint"])
		}
	} else {
		argums = []string{
			"file",
			"new",
			"--version=" + f["version"],
			"--qdir=" + f["qdir"],
			"--cwd=" + cwd,
		}
		argums = append(argums, args...)
	}

	sout, _, _ := qutil.QtechNG(argums, "$..DATA", false, Fcwd)
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
	if len(args) == 0 {
		return "stop"
	}
	return "new"
}

func handleSearch(ui lorca.UI, guiFiller *GuiFiller) string {
	f := make(map[string]string)
	for _, key := range []string{"qpattern", "version", "needle", "perline", "tolower", "regexp", "checkout", "yaml", "jsonpath", "editlist", "clear"} {
		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
		f[key] = value
	}
	guiFiller.Vars = f
	storeVars("search", *guiFiller)

	if f["qpattern"] == "" || f["version"] == "" {
		return "search"
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
	sout, serr, err := qutil.QtechNG(argums, f["jsonpath"], f["yaml"] == "1", Fcwd)
	ui.Eval(`document.getElementById("busy").innerHTML = ""`)
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
	return "search"
}

func handleCheckout(ui lorca.UI, guiFiller *GuiFiller, args []string) string {
	f := make(map[string]string)
	for _, key := range []string{"qpattern", "version", "mode", "yaml", "jsonpath", "editlist", "clear"} {
		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
		f[key] = value
	}
	guiFiller.Vars = f
	storeVars("checkout", *guiFiller)

	if f["qpattern"] == "" || f["version"] == "" {
		return "search"
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
	sout, serr, err := qutil.QtechNG(argums, f["jsonpath"], f["yaml"] == "1", Fcwd)
	ui.Eval(`document.getElementById("busy").innerHTML = ""`)
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
	return "checkout"
}

func handleProperty(ui lorca.UI, guiFiller *GuiFiller) string {
	fname := ui.Eval(`document.getElementById('fname').value`).String()
	clip := ui.Eval(`document.forms[0].clip.value`).String()
	argums := []string{
		"file",
		"tell",
		fname,
		"--tell=" + clip,
	}
	toclip, _, _ := qutil.QtechNG(argums, "", true, Fcwd)

	if toclip != "" {
		argums := []string{
			"clipboard",
			"set",
			toclip,
		}
		qutil.QtechNG(argums, "", true, Fcwd)
	}
	return "stop"
}

func hintoptions(mydir string) string {
	supportdir := qregistry.Registry["qtechng-support-dir"]
	if supportdir == "" {
		return ""
	}
	profiles := path.Join(supportdir, "profiles", "profiles.json")
	data, err := qfs.Fetch(profiles)
	if err != nil {
		return ""
	}
	opt := make([]map[string]string, 0)
	err = json.Unmarshal(data, &opt)
	if err != nil {
		return ""
	}
	options := make([]string, 0)
	options = append(options, "<option selected value=''></option>")
	for _, option := range opt {
		comment := option["comment"]
		if comment == "" {
			continue
		}
		hint := option["hint"]
		if hint == "" {
			hint = "*"
		}
		fname := path.Join(mydir, comment)

		if qfs.IsFile(fname) {
			continue
		}

		options = append(options, "<option value='"+hint+"'>"+comment+"</option>")
	}
	return strings.Join(options, "\n")

}
