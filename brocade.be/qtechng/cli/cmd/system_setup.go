package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qutil "brocade.be/qtechng/lib/util"
)

var systemSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup registry workstation",
	Long: `Setup the registry on your workstation.

The only argument is your user identification on dev.anet.be`,
	Args:    cobra.MinimumNArgs(1),
	Example: "  qtechng system setup rphilips\n  qtechng system setup rphilips /home/rphilips/.ssh/id_rsa",
	RunE:    systemSetup,
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
		log.Fatal("Works only on workstations!")
	}

	xuser := qregistry.Registry["qtechng-user"]
	user := args[0]
	if user == "" || (xuser != "" && xuser != user) {
		log.Fatalf("Identification given (`%s`) does not match existing value(`%s`)", user, xuser)
	}

	qregistry.SetRegistry("qtechng-type", "W")
	qregistry.SetRegistry("qtechng-user", user)
	qregistry.SetRegistry("qtechng-test", "test-entry")
	qregistry.SetRegistry("qtechng-max-parallel", "4")
	qregistry.SetRegistry("os", runtime.GOOS)
	qregistry.SetRegistry("os-sep", string(os.PathSeparator))
	qregistry.SetRegistry("qtechng-exe", "qtechng")
	qregistry.InitRegistry("qtechng-workstation-introspect", "5")

	if runtime.GOOS == "windows" {
		qregistry.SetRegistry("qtechng-exe", "qtechng.exe")
	}

	// bindir
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

	// scratchdir
	homedir, _ := os.UserHomeDir()
	if qregistry.Registry["scratch-dir"] == "" || !qfs.IsDir(qregistry.Registry["scratch-dir"]) {
		scratch := os.TempDir()
		if scratch == "" || !qfs.IsDir(scratch) {
			scratch := path.Join(homedir, "brocade", "tmp")
			os.MkdirAll(scratch, 0700)
		}
		qregistry.SetRegistry("scratch-dir", scratch)
	}

	// several
	qregistry.InitRegistry("qtechng-server", "dev.anet.be:22")
	if qregistry.Registry["qtechng-support-dir"] == "" {
		support := path.Join(homedir, "brocade", "support")
		os.MkdirAll(support, 0700)
		qregistry.SetRegistry("qtechng-support-dir", support)
	}

	qregistry.InitRegistry("qtechng-support-project", "/qtechng/support")
	qregistry.InitRegistry("qtechng-version", "0.00")

	if qregistry.Registry["qtechng-work-dir"] == "" {
		work := path.Join(homedir, "brocade", "source", "data")
		os.MkdirAll(work, 0700)
		qregistry.SetRegistry("qtechng-work-dir", work)
	}

	qregistry.InitRegistry("ssh-default-host", qregistry.Registry["qtechng-server"])
	qregistry.InitRegistry("ssh-default-user", qregistry.Registry["qtechng-user"])

	if len(args) >= 2 && args[1] != "" && qfs.IsFile(args[1]) {
		qregistry.InitRegistry("ssh-default-privatekey", args[1])
	}

	// code
	code, err := exec.LookPath("code")
	if err == nil && code != "" {
		qregistry.InitRegistry("vscode-exe", code)
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

	// releases
	sout, _, _ := qutil.QtechNG([]string{"system", "info", "--remote"}, "$..releases", false)
	if sout != "" {
		x := ""
		err := json.Unmarshal([]byte(sout), &x)
		if err == nil && x != "" {
			qregistry.SetRegistry("qtechng-releases", x)
		}

	}

	return nil
}
