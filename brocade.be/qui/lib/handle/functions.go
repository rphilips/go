package handle

import (
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"

	registry "brocade.be/base/registry"
)

func About(r *http.Request, keys Keys) (Keys, error) {

	cmd := []string{"about"}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func CheckIn(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	cmd := []string{"file", "ci", path}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func CheckOut(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	qpath := strings.Split(path, keys.Workdir)[1]
	cmd := []string{"source", "co", qpath, "--auto"}

	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func Open(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	// to do check if file exists

	cmd := exec.Command(registry.Registry["qtechng-editor-exe"], path)
	err := cmd.Run()
	if err != nil {
		keys.Qresponse = "error" // to do: error handling
	}

	keys.Qresponse = "Opening file: " + path

	return keys, nil
}

func Touch(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	cmd := []string{"fs", "touch", path}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func Tell(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	cmd := []string{"file", "tell", path}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func Previous(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	version := "5.60" // to do get from registry
	qpath := strings.Split(path, keys.Workdir)[1]
	base := filepath.Base(path)
	currdir := strings.Split(path, base)[0]
	prevdir := filepath.Join(currdir, version)
	prevpath := filepath.Join(prevdir, base)

	err := qfs.Mkdir(prevdir, "process")
	if err != nil {
		keys.Qresponse = "error" // to do: error handling
	}

	coCmds := []string{"source", "co", qpath, "--version=" + version, "--cwd=" + prevdir}
	keys.Qresponse = Qcmd(coCmds)

	diff := registry.Registry["qtechng-diff-exe"]
	diff = strings.ReplaceAll(diff, "{target}", path)
	diff = strings.ReplaceAll(diff, "{source}", prevpath)

	var diffCmds []string
	err = json.Unmarshal([]byte(diff), &diffCmds)
	if err != nil {
		keys.Qresponse = "error" // to do: error handling
	}

	cmd := exec.Command(diffCmds[0], diffCmds[1:]...)
	err = cmd.Run()
	if err != nil {
		keys.Qresponse = "error" // to do: error handling
	}

	return keys, nil
}

func Setup(r *http.Request, keys Keys) (Keys, error) {

	cmd := []string{"system", "setup", registry.Registry["qtechng-user"]}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func Commands(r *http.Request, keys Keys) (Keys, error) {

	cmd := []string{"command", "list"}
	keys.Qresponse = Qcmd(cmd)

	return keys, nil
}

func Git(r *http.Request, keys Keys) (Keys, error) {

	path := r.FormValue("path")
	qpath := strings.Split(path, keys.Workdir)[1]
	link := strings.ReplaceAll(registry.Registry["qtechng-vc-url"], "{qpath}", qpath)
	keys.Qresponse = "Version control at: " + link + "<br>"

	return keys, nil
}

func Registry(r *http.Request, keys Keys) (Keys, error) {

	path := registry.Registry["brocade-registry-file"]

	cmd := exec.Command(registry.Registry["qtechng-editor-exe"], path)
	err := cmd.Run()
	if err != nil {
		keys.Qresponse = "error" // to do: error handling
	}

	keys.Qresponse = "Opening file: " + path

	return keys, nil
}
