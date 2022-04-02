package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
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
	Long:    `All kinds of functions for the qtechng Graphical User Interface`,
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
	if Fmenu == "checkout" && len(args) != 0 {
		Fmenu = "checkoutargs"
	}
	standalone := Fmenu != "menu" && Fmenu != ""
	width := 520
	height := 780

	switch Fmenu {
	case "property":
		width = 680
	case "diff":
		if len(args) == 0 {
			return nil
		}
		myfile := qutil.AbsPath(args[0], Fcwd)
		if !qfs.IsFile(myfile) {
			return nil
		}
		args = []string{myfile}
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
				case "touch":
					sout := handleTouch(Fcwd, args)
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
				case "checkoutargs":
					sout := handleCheckoutArgs(Fcwd, args)
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
						out, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, Fcwd)
						guiFiller.Vars["properties"] = string(out)
					}
				case "new":
					mydir := Fcwd
					argums := []string{
						"dir",
						"tell",
						"--cwd=" + mydir,
					}
					out, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, Fcwd)
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
				case "diff":
					mydir := Fcwd
					myfile := args[0]
					argums := []string{
						"file",
						"tell",
						myfile,
						"--cwd=" + mydir,
					}
					out, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, Fcwd)
					m := make(map[string]string)
					json.Unmarshal([]byte(out), &m)
					qpath := m["qpath"]
					version := m["version"]
					if version == "" {
						version = "0.00"
					}
					guiFiller.Vars["qpath"] = qpath
					guiFiller.Vars["version"] = version
					guiFiller.Vars["name"] = myfile
					guiFiller.VarsS = make(map[string]template.HTML)
					guiFiller.VarsS["select"] = template.HTML(diffoptions(qpath, version))
				}

				t, err := template.ParseFS(guifs, "templates/"+Fmenu+".html")
				if err != nil {
					Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
					return
				}
				buf := new(bytes.Buffer)
				err = t.Execute(buf, guiFiller)

				if err != nil {
					Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
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
				case "touch":
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
				case "checkoutargs":
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

				case "touch":
					ui.Bind("golangfunc", func(indicator string) {
						menulistener <- "stop"
					})
				case "checkoutargs":
					ui.Bind("golangfunc", func(indicator string) {
						menulistener <- "stop"
					})

				case "search":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						//handleSearch(ui, &guiFiller, "search")
						menulistener <- handleSearch(ui, &guiFiller, "search")
					})
				case "checkout":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						menulistener <- handleSearch(ui, &guiFiller, "checkout")
						// if len(args) == 0 {
						// 	menulistener <- handleSearch(ui, &guiFiller, "checkout")
						// } else {
						// 	menulistener <- handleCheckout(ui, &guiFiller, args)
						// }
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

				case "diff":
					ui.Bind("golangfunc", func(indicator string) {
						if indicator == "stop" {
							menulistener <- "stop"
							return
						}
						menulistener <- handleDiff(ui, &guiFiller, Fcwd)
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
	fname := filepath.Join(qregistry.Registry["scratch-dir"], "menu-"+menu+".json")
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
	if vars["tree"] != "1" && vars["flat"] != "1" {
		vars["auto"] = "1"
	}
	for _, k := range []string{"tolower", "regexp", "checkout", "clear", "auto", "flat", "tree"} {
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
	fname := filepath.Join(qregistry.Registry["scratch-dir"], "menu-"+menu+".json")

	m := make(map[string]string)
	for k, v := range guiFiller.Vars {
		if k == "checkout" && menu == "search" {
			continue
		}
		m[k] = v
	}
	b, err := json.Marshal(m)
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
	sout, _, _ := qutil.QtechNG(argums, []string{"$..file"}, true, cwd)
	return sout
}

func handleTouch(cwd string, args []string) string {
	argums := []string{
		"fs",
		"touch",
		"--cwd=" + cwd,
		"--recurse",
	}
	if len(args) != 0 {
		argums = append(argums, args...)
	}
	sout, _, _ := qutil.QtechNG(argums, []string{"$..touched"}, true, cwd)
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
	var argums []string
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

	sout, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, Fcwd)
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

func handleSearch(ui lorca.UI, guiFiller *GuiFiller, sech string) string {
	f := make(map[string]string)
	for _, key := range []string{"qpattern", "version", "needle", "perline", "tolower", "regexp", "checkout", "yaml", "jsonpath", "editlist", "clear"} {
		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
		f[key] = value
	}
	guiFiller.Vars = f
	storeVars(sech, *guiFiller)

	if f["qpattern"] == "" || f["version"] == "" {
		return sech
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
	sout, serr, err := qutil.QtechNG(argums, []string{f["jsonpath"]}, f["yaml"] == "1", Fcwd)
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
	return sech
}

// func handleCheckout(ui lorca.UI, guiFiller *GuiFiller, args []string) string {
// 	f := make(map[string]string)
// 	for _, key := range []string{"qpattern", "version", "yaml", "jsonpath", "editlist", "clear", "checkout"} {
// 		value := ui.Eval(`document.getElementById('` + key + `').value`).String()
// 		f[key] = value
// 	}
// 	for _, mode := range []string{"auto", "tree", "flat"} {
// 		value := ui.Eval(`document.getElementById('mode_` + mode + `').checked ? '1' : '0'`).String()
// 		if value == "1" {
// 			f["mode"] = mode
// 			break
// 		}
// 	}
// 	guiFiller.Vars = f
// 	storeVars("checkout", *guiFiller)

// 	if f["qpattern"] == "" || f["version"] == "" {
// 		return "search"
// 	}
// 	argums := []string{
// 		"source",
// 	}

// 	if f["checkout"] == "1" {
// 		mode := "--" + f["mode"]
// 		if mode == "--flat" {
// 			mode = ""
// 		}
// 		argums = append(argums, "co")
// 		if mode != "" {
// 			argums = append(argums, mode)
// 		}
// 		if f["clear"] == "1" {
// 			argums = append(argums, "--clear")
// 		}
// 	} else {
// 		argums = append(argums, "list")
// 	}
// 	argums = append(argums, "--version="+f["version"])
// 	argums = append(argums, "--qpattern="+f["qpattern"])
// 	argums = append(argums, "--needle="+f["needle"])
// 	argums = append(argums, "--list="+f["editlist"])
// 	if f["perline"] == "1" {
// 		argums = append(argums, "--perline")
// 	}
// 	if f["tolower"] == "1" {
// 		argums = append(argums, "--tolower")
// 	}
// 	if f["regexp"] == "1" {
// 		argums = append(argums, "--regexp")
// 	}

// 	ui.Eval(`document.getElementById("busy").innerHTML = "Busy ..."`)
// 	ui.Eval(`document.getElementById("busy").style="display:block;")`)
// 	sout, serr, err := qutil.QtechNG(argums, f["jsonpath"], f["yaml"] == "1", Fcwd)
// 	ui.Eval(`document.getElementById("busy").innerHTML = ""`)
// 	if err != nil {
// 		serr += "\n\nError:" + err.Error()
// 	}
// 	bx, _ := json.Marshal(sout)
// 	sx := string(bx)
// 	st := ""
// 	if len(sx) > 2 {
// 		st = sx[:2]
// 	}
// 	k := strings.IndexAny(st, "[{")
// 	if k == 1 {
// 		ui.Eval(`document.getElementById("jsondisplay").innerHTML = syntaxHighlight(` + sx + `)`)
// 		ui.Eval(`document.getElementById("yamldisplay").innerHTML = ""`)
// 	} else {
// 		ui.Eval(`document.getElementById("yamldisplay").innerHTML = ` + sx)
// 		ui.Eval(`document.getElementById("jsondisplay").innerHTML = ""`)
// 	}
// 	return "checkout"
// }

func handleCheckoutArgs(cwd string, args []string) string {
	files := make([]string, 0)
	dirs := make([]string, 0)
	for _, arg := range args {
		fmt.Println("arg:", arg)
		arg := qutil.AbsPath(arg, cwd)
		if qfs.IsFile(arg) {
			files = append(files, arg)
			continue
		}
		if qfs.IsDir(arg) {
			fmt.Println("arg2:", arg)
			dirs = append(dirs, arg)
		}
	}

	qpaths := make([]string, 0)

	if len(files) != 0 {
		argums := []string{"file", "refresh"}
		argums = append(argums, files...)
		sout, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, cwd)
		if sout != "" {
			slice := make([]string, 0)
			e := json.Unmarshal([]byte(sout), &slice)
			if e == nil {
				qpaths = append(qpaths, slice...)
			}
		}

	}
	if len(dirs) != 0 {
		argums := []string{"dir", "refresh"}
		argums = append(argums, dirs...)
		sout, _, _ := qutil.QtechNG(argums, []string{"$..DATA"}, false, cwd)
		if sout != "" {
			slice := make([]string, 0)
			e := json.Unmarshal([]byte(sout), &slice)
			if e == nil {
				qpaths = append(qpaths, slice...)
			}
		}
	}
	argums := []string{"file", "list", "--recurse"}
	for _, qp := range qpaths {
		argums = append(argums, "--qpattern="+qp)
	}

	sout, _, _ := qutil.QtechNG(argums, []string{"$..qpath"}, true, cwd)

	return sout
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
	toclip, _, _ := qutil.QtechNG(argums, nil, true, Fcwd)

	if toclip != "" {
		argums := []string{
			"clipboard",
			"set",
			toclip,
		}
		qutil.QtechNG(argums, nil, true, Fcwd)
	}
	return "stop"
}

func handleDiff(ui lorca.UI, guiFiller *GuiFiller, cwd string) string {

	diff := qregistry.Registry["qtechng-diff-exe"]

	if diff == "" {
		return "stop"
	}

	cversion := ui.Eval(`document.getElementById('cversion').value`).String()
	file := ui.Eval(`document.getElementById('name').value`).String()
	qpath := ui.Eval(`document.getElementById('qpath').value`).String()
	version := ui.Eval(`document.getElementById('version').value`).String()

	targetdir := filepath.Dir(file)
	source := file
	target := file
	if version == cversion {
		source = file + ".ori"
		target = file
		qfs.CopyFile(target, source, "qtech", false)
	} else {
		source = file
		targetdir = filepath.Join(filepath.Dir(source), cversion)
		target = filepath.Join(targetdir, filepath.Base(file))
	}
	args := []string{
		"source",
		"co",
		qpath,
		"--version=" + cversion,
	}

	qutil.QtechNG(args, nil, false, targetdir)
	if !qfs.IsFile(source) {
		return "stop"
	}
	if !qfs.IsFile(target) {
		return "stop"
	}
	exe := make([]string, 0)

	json.Unmarshal([]byte(diff), &exe)

	if len(exe) < 2 {
		return "stop"
	}

	pexe, _ := exec.LookPath(exe[0])
	argums := make([]string, 0)
	for _, arg := range exe {
		arg = strings.ReplaceAll(arg, "{source}", source)
		arg = strings.ReplaceAll(arg, "{target}", target)
		argums = append(argums, arg)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   pexe,
		Args:   argums,
		Dir:    targetdir,
		Stdout: &stdout,
		Stderr: &stderr,
	}
	cmd.Run()
	return "stop"
}

func hintoptions(mydir string) string {
	supportdir := qregistry.Registry["qtechng-support-dir"]
	if supportdir == "" {
		return ""
	}
	profiles := filepath.Join(supportdir, "profiles", "profiles.json")
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
		fname := filepath.Join(mydir, comment)

		if qfs.IsFile(fname) {
			continue
		}

		options = append(options, "<option value='"+hint+"'>"+comment+"</option>")
	}
	return strings.Join(options, "\n")

}

func diffoptions(myfile string, version string) string {
	prevs := qregistry.Registry["qtechng-releases"]
	prevs = strings.ReplaceAll(prevs, ",", " ")
	prevs = strings.ReplaceAll(prevs, ";", " ")
	prevs = strings.ReplaceAll(prevs, "\t", " ")
	options := make([]string, 0)
	for _, v := range strings.SplitN(prevs, " ", -1) {
		if v == "" {
			continue
		}
		if v == version {
			break
		}
		options = append(options, "<option value='"+v+"'>"+v+"</option>")
	}
	options = append(options, "<option selected value='"+version+"'>Current in repository</option>")
	for i, j := 0, len(options)-1; i < j; i, j = i+1, j-1 {
		options[i], options[j] = options[j], options[i]
	}
	return strings.Join(options, "\n")
}
