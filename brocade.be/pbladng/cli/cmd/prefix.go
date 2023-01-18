package cmd

import (
	"bytes"
	"embed"
	"encoding/json"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"

	bfs "brocade.be/base/fs"
	ptools "brocade.be/pbladng/lib/tools"

	"github.com/spf13/cobra"
	"github.com/webview/webview"
	"go.lsp.dev/uri"
)

//go:embed templates
var guifsp embed.FS

var prefixCmd = &cobra.Command{
	Use:   "prefix",
	Short: "File manipulation",
	Long:  `Prefix file names interactively`,

	Example: `gopblad prefix`,
	RunE:    prefix,
}

func init() {
	prefixCmd.PersistentFlags().BoolVar(&Fstderr, "stderr", false, "Show stderr")
	rootCmd.AddCommand(prefixCmd)
}

type RawImage struct {
	Name string
	Base string
	Id   string
	URL  string
}
type Dir struct {
	Root   string
	Images []RawImage
}

func prefix(cmd *cobra.Command, args []string) error {
	if !Fstderr {
		_, err := ptools.Launch([]string{"gopblad", "prefix", "--stderr", "--cwd=."}, nil, Fcwd, false, false)
		return err
	}
	if len(args) == 0 {
		args = []string{`*.[Jj][Pp][Gg]`, `*.[Jj][Pp][Ee][Gg]`}
	}

	// HTML

	files, _, err := bfs.FilesDirs(Fcwd)
	if err != nil {
		return err
	}

	for _, z := range files {
		fname := strings.ToLower(z.Name())
		if strings.HasSuffix(fname, ".zip") {
			err := unzip(fname, Fcwd)
			if err != nil {
				return err
			}
		}
	}
	files, _, err = bfs.FilesDirs(Fcwd)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	work := make([]os.FileInfo, 0, len(files))
	for _, z := range files {
		fname := z.Name()
		if len(args) != 0 {
			for _, a := range args {
				if filepath.Base(a) == fname {
					work = append(work, z)
					break
				}
				ok, e := filepath.Match(a, fname)
				if e != nil {
					return e
				}
				if ok {
					work = append(work, z)
					break
				}
			}
		} else {
			work = append(work, z)
		}
	}

	if len(work) == 0 {
		return nil
	}

	images := make([]RawImage, 0, len(work))

	for _, w := range work {
		img := RawImage{
			Name: w.Name(),
			Base: strings.TrimSuffix(w.Name(), filepath.Ext(w.Name())),
			Id:   "",
			URL:  string(uri.File(filepath.Join(Fcwd, w.Name()))),
		}
		images = append(images, img)
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Name < images[j].Name })
	for i := range images {
		images[i].Id = "img" + strconv.Itoa(i)
	}

	guiFiller := Dir{
		Root:   Fcwd,
		Images: images,
	}

	t, err := template.ParseFS(guifsp, "templates/prefix.html")
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, guiFiller)

	if err != nil {
		return err
	}

	// interface
	guilistener := make(chan string)
	defer close(guilistener)
	os.Stderr, _ = os.Open(os.DevNull)
	go func() {
		os.Stderr, _ = os.Open(os.DevNull)
		width := 720
		height := 780

		ui := webview.New(false)
		defer ui.Destroy()

		ui.SetSize(width, height, webview.HintNone)
		ui.SetHtml(buf.String())
		ui.Bind("register", func(action string) {
			handleForm(action, ui, guilistener)
		})
		ui.Run()
	}()
	<-guilistener

	return nil
}

func handleForm(action string, ui webview.WebView, guilistener chan string) {
	info := make(map[string]string)
	json.Unmarshal([]byte(action), &info)
	renames := make(map[string]string)
	pcount := make(map[string]int)
	i := -1
	for {
		i++
		key := "img" + strconv.Itoa(i)
		fname := info[key+"-old"]
		if fname == "" {
			break
		}
		prefix := info[key+"-prefix"]
		newname := info[key+"-new"]

		pcount[prefix] = 1 + pcount[prefix]
		s := strings.ReplaceAll(prefix, "#", strconv.Itoa(pcount[prefix]))
		s = strings.ReplaceAll(s, "@", newname)

		if s == "" {
			continue
		}
		ext := filepath.Ext(fname)
		newname = strings.TrimSuffix(s, ext)
		if newname == "" {
			continue
		}
		ext = strings.ToLower(ext)
		if ext == ".jpeg" {
			ext = ".jpg"
		}
		newname += ext
		olddir := filepath.Dir(fname)
		oldbase := filepath.Base(fname)
		newbase := filepath.Base(newname)
		if oldbase == newbase {
			continue
		}
		renames[fname] = filepath.Join(olddir, newbase)
	}
	switch info["action"] {

	case "stop":
		ui.Eval("window.close()")
		ui.Terminate()
		guilistener <- `{"action": "close"}`
		return
	case "rename":
		msg := ""
		if len(renames) == 0 {
			msg := "<p>Nothing to rename!</p>"
			ui.Eval("Rename('" + msg + "')")
			return
		}
		msg = "<table>"
		fnames := make([]string, 0, len(renames))
		for f := range renames {
			fnames = append(fnames, f)
		}
		sort.Strings(fnames)
		for _, f := range fnames {
			msg += "<tr><td><var>" + html.EscapeString(f) + "</var></td><td>&#10230;</td><td><var>" + html.EscapeString(renames[f]) + "</var></td><tr>"
		}
		msg += "<tr><td></td><td></td><td><input class=\"label\" type=\"submit\" value=\"Confirm Rename\" onclick=\"register(makeJSON())\" /></td><tr>"
		msg += "</table>"
		ui.Eval("Rename('" + msg + "')")
		return
	case "confirm":
		dorename(renames, nil)
		ui.Eval("window.close()")
		ui.Terminate()
		guilistener <- `{"action": "close"}`
	}

}
