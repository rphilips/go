package handle

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	registry "brocade.be/base/registry"
	html "brocade.be/qui/lib/html"
	util "brocade.be/qui/lib/util"
)

const port = ":8081"
const baseURL = "http://localhost" + port + "/"

type Keys struct {
	BaseURL   string
	Name      string
	Qfiles    []string
	Qpaths    []string
	Workdir   string
	Qresponse string
}

// Handler function for start screen
func Start(w http.ResponseWriter, r *http.Request) {
	var keys Keys
	workdir := registry.Registry["qtechng-work-dir"]

	fn := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".qtechng" {
			return nil
		}
		if strings.Contains(path, ".vscode") {
			return nil
		}
		// no new variables for performance
		qpath := strings.Split(path, workdir)[1]
		keys.Qpaths = append(keys.Qpaths, (`<span style="cursor:pointer;" onclick="document.getElementById('file').value='` +
			path + `';document.getElementById('path').value='` + qpath + `'">` + qpath + `</span><br>`))
		keys.Qfiles = append(keys.Qfiles, path)
		return nil
	}

	err := filepath.WalkDir(workdir, fn)
	if err != nil {
		fmt.Println(err)
	}

	start := html.Start(keys)
	fmt.Fprintln(w, start)
}

// Handler function for rename screen
func Rename(w http.ResponseWriter, r *http.Request) {
	var keys Keys
	keys.BaseURL = baseURL

	rename := html.Rename(keys)
	fmt.Fprintln(w, rename)
}

// Handler function for delete screen
func Delete(w http.ResponseWriter, r *http.Request) {
	var keys Keys
	keys.BaseURL = baseURL

	rename := html.Delete(keys)
	fmt.Fprintln(w, rename)
}

// Handler function for result screen
func Result(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		fmt.Fprintf(w, "ERROR parsing: %v", err)
		return
	}

	var keys Keys
	keys.BaseURL = baseURL
	keys.Workdir = registry.Registry["qtechng-work-dir"]

	err = nil

	switch r.FormValue("cmd") {
	case "about":
		keys, err = About(r, keys)
	case "checkin":
		keys, err = CheckIn(r, keys)
	case "checkout":
		keys, err = CheckOut(r, keys)
	case "open":
		keys, err = Open(r, keys)
	case "touch":
		keys, err = Touch(r, keys)
	case "registry":
		keys, err = Registry(r, keys)
	case "tell":
		keys, err = Tell(r, keys)
	case "setup":
		keys, err = Setup(r, keys)
	case "commands":
		keys, err = Commands(r, keys)
	case "git":
		keys, err = Git(r, keys)
	case "previous":
		keys, err = Previous(r, keys)
	case "all":
		keys, err = All(r, keys)
	case "quit":
		keys, err = Quit(r, keys)
	}

	if err != nil {
		fmt.Fprintf(w, "ERROR: %v", err)
		return
	}

	keys.Qresponse = util.ToHTML(keys.Qresponse)
	result := html.Result(keys)
	fmt.Fprint(w, result)
}
