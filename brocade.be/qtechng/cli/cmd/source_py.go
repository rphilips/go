package cmd

import (
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourcePyCmd = &cobra.Command{
	Use:   "py",
	Short: "Execute a Python script in the repository",
	Long: `This command retrieves the content of the corresponding qtechng source file
in a temporary directory and executes the Python script in this directory.

On workstations, use *qtechng file py*
`,
	Example: "qtechng source py /core/qtech/local.py",
	Args:    cobra.MinimumNArgs(1),
	RunE:    sourcePy,

	Annotations: map[string]string{
		"with-qtechtype": "BP",
		"fill-version":   "yes",
	},
}

// Fpyonly Python exe only
var Fpyonly bool = false

func init() {
	sourcePyCmd.Flags().BoolVar(&Fpyonly, "pyonly", false, "Return Python executable")
	sourcePyCmd.Flags().StringVar(&Fpy, "py", "", "Python default executable (py2 | py3")
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
			Msg:   []string{"Script is a qpath and should start with `/`"},
		}
		return e
	}
	argums := []string{
		"source",
		"co",
		pyscript,
		"--neighbours",
		"--tree",
		"--version=" + Fversion,
	}
	qutil.QtechNG(argums, nil, false, tmpdir)

	parts := strings.SplitN(pyscript, "/", -1)
	parts[0] = tmpdir
	script := filepath.Join(parts...)
	args[0] = script
	//fmt.Println("errors:", err, stdout, stderr, argums, tmpdir, args)
	return filePy(cmd, args)
}
