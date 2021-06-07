package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
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
	Short: "Setup registry workstation",
	Long:  `Setup the registry on your workstation`,
	Args:  cobra.MaximumNArgs(1),
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

	qt := qregistry.Registry["qtechng-type"]
	if qt != "" && qt != "W" {
		Fmsg = qreport.Report("", errors.New("works only on workstations"), Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}

	if len(args) == 1 {
		qregistry.SetRegistry("qtechng-user", args[0])
		qregistry.SetRegistry("ssh-default-user", args[0])
	}

	// base
	qregistry.SetRegistry("qtechng-type", "W")
	qregistry.SetRegistry("qtechng-test", "test-entry")
	qregistry.SetRegistry("qtechng-max-parallel", "4")
	qregistry.SetRegistry("os", runtime.GOOS)
	qregistry.SetRegistry("os-sep", string(os.PathSeparator))
	qregistry.InitRegistry("qtechng-workstation-introspect", "5")
	qregistry.InitRegistry("qtechng-support-project", "/qtechng/support")
	qregistry.InitRegistry("qtechng-version", "0.00")

	if runtime.GOOS == "windows" {
		qregistry.SetRegistry("qtechng-exe", "qtechng.exe")
	} else {
		qregistry.SetRegistry("qtechng-exe", "qtechng")
	}
	qregistry.InitRegistry("qtechng-server", "dev.anet.be:22")
	qregistry.InitRegistry("ssh-default-host", qregistry.Registry["qtechng-server"])
	server := strings.SplitN(qregistry.Registry["qtechng-server"]+":", ":", -1)[0]
	url := fmt.Sprintf("https://%s/qtechng/qtechng-%s-%s", server, runtime.GOOS, runtime.GOARCH)
	qregistry.SetRegistry("qtechng-url", url)

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

	// get 'about'
	soutr, serrr, err := qutil.QtechNG([]string{"about", "--remote"}, "$..DATA", false, Fcwd)
	if err != nil {
		Fmsg = qreport.Report(serrr, err, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}
	mr := make(map[string]string)
	err = json.Unmarshal([]byte(soutr), &mr)
	if err != nil {
		Fmsg = qreport.Report(soutr, err, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}
	username := mr["!!user.username"]
	if username != "" {
		qregistry.InitRegistry("qtechng-user", username)
	}
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

	// QtechNG
	soutl, serrl, err := qutil.QtechNG([]string{"about"}, "$..DATA", false, Fcwd)
	if err != nil {
		Fmsg = qreport.Report(serrl, err, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}
	ml := make(map[string]string)
	err = json.Unmarshal([]byte(soutl), &ml)
	if err != nil {
		Fmsg = qreport.Report(soutl, err, Fjq, Fyaml, Funquote, Fsilent)
		return nil
	}
	// releases
	sout, _, _ := qutil.QtechNG([]string{"system", "info", "--remote"}, "$..releases", false, Fcwd)
	if sout != "" {
		x := ""
		err := json.Unmarshal([]byte(sout), &x)
		if err == nil && x != "" {
			qregistry.SetRegistry("qtechng-releases", x)
		}
	}

	if ml["..!BuildTime"] != mr["..!BuildTime"] {
		qutil.RefreshBinary()

	}
	return nil
}
