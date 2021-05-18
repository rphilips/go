package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var sourcePyCmd = &cobra.Command{
	Use:   "py",
	Short: "Executes a python script in the repository",
	Long: `Retrieves the content of the corresponding qtech repository
			  in a temporary directory and executes the python script in this
			  directory`,
	Example: "qtechng source py /core/qtech/local.py",
	Args:    cobra.MinimumNArgs(1),
	RunE:    sourcePy,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"fill-version":   "yes",
	},
}

// Fpyonly Python exe only
var Fpyonly bool = false

func init() {
	sourcePyCmd.Flags().BoolVar(&Fpyonly, "pyonly", false, "return python executable")
	sourceCmd.AddCommand(sourcePyCmd)
}

func sourcePy(cmd *cobra.Command, args []string) error {
	tmpdir, err := qfs.TempDir("", "py")
	if err != nil {
		return err
	}
	pyscript := args[0]
	if !strings.HasSuffix(pyscript, ".py") {
		e := &qerror.QError{
			Ref:   []string{errRoot + "py"},
			QPath: pyscript,
			Msg:   []string{"Script should end with `.py`"},
		}
		return e
	}
	if !strings.HasPrefix(pyscript, "/") {
		e := &qerror.QError{
			Ref:   []string{errRoot + "py"},
			QPath: pyscript,
			Msg:   []string{"Script should start with `/`"},
		}
		return e
	}

	executable, err := os.Executable()
	prog := filepath.Base(os.Args[0])
	if err != nil {
		if prog == os.Args[0] {
			executable, _ = exec.LookPath(prog)
		} else {
			executable = os.Args[0]
		}
	}

	argums := []string{
		prog,
		"source",
		"co",
		pyscript,
		"--neighbours",
		"--tree",
		"--version=" + Fversion,
		"--silent",
	}

	qcmd := exec.Cmd{
		Path: executable,
		Args: argums,
		Dir:  tmpdir,
	}

	qcmd.Run()

	parts := strings.SplitN(pyscript, "/", -1)
	parts[0] = tmpdir
	script := filepath.Join(parts...)
	args[0] = script
	return filePy(cmd, args)
}
