package cmd

import (
	"archive/tar"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
)

var systemSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup the registry",
	Long: `This command sets up basic registry values for qtechng.
The (optional) argument provided is the *qtechng-user* value.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  qtechng system setup
    qtechng system setup rphilips
`,
	RunE: systemSetup,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	systemCmd.AddCommand(systemSetupCmd)
}

func systemSetup(cmd *cobra.Command, args []string) error {
	if qregistry.Registry["error"] != "" {
		Fmsg = qreport.Report(nil, errors.New(qregistry.Registry["error"]), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	qt := qregistry.Registry["qtechng-type"]
	onW := qt == "" || qt == "W"
	onB := strings.ContainsRune(qt, 'B')

	if onW && len(args) == 1 {
		qregistry.SetRegistry("qtechng-user", args[0])
		qregistry.SetRegistry("ssh-default-user", args[0])
	}

	// base
	qregistry.SetRegistry("qtechng-test", "test-entry")
	qregistry.SetRegistry("qtechng-vc-url", "https://dev.anet.be/cgit/cgit.cgi/qtechng/tree/data{qpath}")

	qregistry.InitRegistry("qtechng-server", "dev.anet.be:22")
	qregistry.InitRegistry("ssh-default-host", qregistry.Registry["qtechng-server"])
	server := strings.SplitN(qregistry.Registry["qtechng-server"]+":", ":", -1)[0]
	url := fmt.Sprintf("https://%s/qtechng/qtechng-%s-%s", server, runtime.GOOS, runtime.GOARCH)
	qregistry.SetRegistry("qtechng-url", url)
	if onW {
		qregistry.SetRegistry("qtechng-type", "W")
		qregistry.SetRegistry("qtechng-max-parallel", "4")
		qregistry.SetRegistry("os", runtime.GOOS)
		qregistry.SetRegistry("os-sep", string(os.PathSeparator))
		qregistry.InitRegistry("qtechng-workstation-introspect", "5")
		qregistry.InitRegistry("qtechng-version", "0.00")
		if runtime.GOOS == "windows" {
			qregistry.SetRegistry("qtechng-exe", "qtechng.exe")
		} else {
			qregistry.SetRegistry("qtechng-exe", "qtechng")
		}
		if qregistry.Registry["bindir"] == "" {
			ex, err := os.Executable()
			if err != nil {
				ex = ""
			}
			if ex != "" {
				dir, err := filepath.Abs(filepath.Dir(ex))
				if err == nil {
					qregistry.SetRegistry("bindir", dir)
				}
			}
		}
	}
	if onB {
		err := qutil.RefreshBinary(true)
		if err != nil {
			Fmsg = qreport.Report("", err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
		return nil
	}

	// get 'about'
	soutr, serrr, err := qutil.QtechNG([]string{"about", "--remote"}, []string{"$..DATA"}, false, Fcwd)
	if err != nil {
		Fmsg = qreport.Report(serrr, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	mr := make(map[string]string)
	err = json.Unmarshal([]byte(soutr), &mr)
	if err != nil {
		Fmsg = qreport.Report(soutr, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	username := mr["!!user.username"]
	if onW && username != "" {
		qregistry.InitRegistry("qtechng-user", username)
	}
	if onW {
		qregistry.InitRegistry("ssh-default-user", qregistry.Registry["qtechng-user"])
		// user related dirs
		homedir, _ := os.UserHomeDir()
		// scratchdir
		if qregistry.Registry["scratch-dir"] == "" || !qfs.IsDir(qregistry.Registry["scratch-dir"]) {
			scratch := os.TempDir()
			if scratch == "" || !qfs.IsDir(scratch) {
				scratch := filepath.Join(homedir, "brocade", "tmp")
				os.MkdirAll(scratch, 0700)
			}
			qregistry.SetRegistry("scratch-dir", scratch)
		}
		// workdir
		if qregistry.Registry["qtechng-work-dir"] == "" {
			work := filepath.Join(homedir, "brocade", "source", "data")
			os.MkdirAll(work, 0700)
			qregistry.SetRegistry("qtechng-work-dir", work)
		}
		// logdir
		if qregistry.Registry["qtechng-log-dir"] == "" {
			logdir := filepath.Join(homedir, "brocade", "log")
			os.MkdirAll(logdir, 0700)
			qregistry.SetRegistry("qtechng-log-dir", logdir)
		}
		// supportdir
		if qregistry.Registry["qtechng-support-dir"] == "" {
			support := filepath.Join(homedir, "brocade", "support")
			os.MkdirAll(support, 0700)
			qregistry.SetRegistry("qtechng-support-dir", support)
		}

		// binaries
		// code
		code, err := exec.LookPath("code")
		if err == nil && code != "" {
			qregistry.InitRegistry("vscode-exe", filepath.Base(code))
		}
		// diff
		meld, err := exec.LookPath("meld")
		if err == nil && meld != "" {
			qregistry.InitRegistry("qtechng-diff-exe", "[\"meld\", \"{target}\", \"{source}\"]")
		}
		winm, err := exec.LookPath("WinMergeU")
		if err == nil && winm != "" {
			qregistry.InitRegistry("qtechng-diff-exe", "[\"WinMergeU\", \"{source}\", \"{target}\"]")
		}
		dm, err := exec.LookPath("DiffMerge")
		if err == nil && dm != "" {
			qregistry.InitRegistry("qtechng-diff-exe", "[\"DiffMerge\", \"{source}\", \"{target}\"]")
		}
		kdm, err := exec.LookPath("kdiff3")
		if err == nil && kdm != "" {
			qregistry.InitRegistry("qtechng-diff-exe", "[\"kdiff3\", \"{source}\", \"{target}\"]")
		}
	}
	// QtechNG
	soutl, serrl, err := qutil.QtechNG([]string{"about"}, []string{"$..DATA"}, false, Fcwd)
	if err != nil {
		Fmsg = qreport.Report(serrl, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	ml := make(map[string]string)
	err = json.Unmarshal([]byte(soutl), &ml)
	if err != nil {
		Fmsg = qreport.Report(soutl, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	// releases
	if onW {
		sout, _, _ := qutil.QtechNG([]string{"system", "info", "--remote"}, []string{"$..releases"}, false, Fcwd)
		if sout != "" {
			x := ""
			err := json.Unmarshal([]byte(sout), &x)
			if err == nil && x != "" {
				qregistry.SetRegistry("qtechng-releases", x)
			}
		}
	}

	err = qutil.RefreshBinary(ml["!BuildTime"] != mr["!BuildTime"] || strings.ContainsAny(qregistry.Registry["qtechng-type"], "BP"))
	if err != nil {
		Fmsg = qreport.Report("", err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	err = systemSupport()
	r := ""
	if err == nil {
		r = "QtechNG installed. Check with `qtechng about`"
	}
	Fmsg = qreport.Report(r, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}

// Untar untars a tarball and writes the files to a target path
// Source: https://golangdocs.com/tar-gzip-in-golang
func Untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

func systemSupport() error {

	qt := qregistry.Registry["qtechng-type"]
	onW := strings.ContainsRune(qt, 'W')
	onB := strings.ContainsRune(qt, 'B')
	onP := strings.ContainsRune(qt, 'P')
	if !onW || onB || onP {
		return nil
	}

	supportDir := qregistry.Registry["qtechng-support-dir"]
	if supportDir == "" {
		return errors.New("installation support: registry qtechng-support-dir is empty or missing")
	}

	qtechngURL := qregistry.Registry["qtechng-url"]
	if qtechngURL == "" {
		return errors.New("installation support: registry qtechng-url is empty or missing")
	}

	qSegments := strings.Split(qtechngURL, "/")
	binary := qSegments[len(qSegments)-1]
	qtechngSubDomain := strings.TrimRight(qtechngURL, binary)
	tarURL := strings.Join([]string{qtechngSubDomain, "support.tar"}, "")

	fileURL, err := url.Parse(tarURL)
	if err != nil {
		return errors.New("installation support: parsing of `" + tarURL + "` fails: " + err.Error())
	}
	path := fileURL.Path
	fileSegments := strings.Split(path, "/")
	tarPath := filepath.Join(supportDir, fileSegments[len(fileSegments)-1])
	err = qfs.GetURL(tarURL, tarPath, "tempfile")
	if err != nil {
		return errors.New("installation support: retrieval of `" + tarURL + "` fails: " + err.Error())
	}

	// Untar file
	err = Untar(tarPath, supportDir)
	if err != nil {
		return errors.New("installation support: untar-ing of `" + tarPath + "` fails: " + err.Error())
	}
	os.Remove(tarPath)

	// install vsix for vscode

	dir := filepath.Join(supportDir, "vscode")
	err = qutil.VSCode(dir)
	if err != nil {
		return errors.New("installation support: unpacking of vscode extensions fails: " + err.Error())
	}
	return nil
}
