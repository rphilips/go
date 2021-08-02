package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qpy "brocade.be/base/python"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

// Fpysource extra arguments
var Fpysource = make([]string, 0)

var filePyCmd = &cobra.Command{
	Use:   "py",
	Short: "Executes a python script in the local filesystem",
	Long: `Executes the python script in the local filesystem.

The system tries to find the appropriate Python interpreter and executes the script.
The first argument is the python script the other arguments are parameters for this
script.
`,
	Example: "qtechng file py /home/rphilips/core/qtech/local.py",
	Args:    cobra.MinimumNArgs(1),
	RunE:    filePy,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"fill-version":   "no",
	},
}

// Fpy pythontype
var Fpy string

func init() {
	filePyCmd.Flags().BoolVar(&Fpyonly, "pyonly", false, "return python executable")
	filePyCmd.Flags().StringVar(&Fpy, "py", "", "Python default executable (py2 | py3")
	fileCmd.AddCommand(filePyCmd)
}

func filePy(cmd *cobra.Command, args []string) error {
	pyscript := args[0]
	if !strings.HasSuffix(pyscript, ".py") {
		e := &qerror.QError{
			Ref:  []string{"file1.py"},
			File: pyscript,
			Msg:  []string{"Script should end with `.py`"},
		}
		return e
	}

	pyscript, _ = qfs.AbsPath(filepath.Join(Fcwd, pyscript))

	py := Fpy
	if py == "" {
		py = qutil.GetPy(pyscript)
	}

	if py == "" {
		e := &qerror.QError{
			Ref:  []string{"file2.py"},
			File: pyscript,
			Msg:  []string{"Cannot determine python executable associated with `" + pyscript + "`"},
		}
		return e
	}

	py3 := py == "py3"

	if Fpyonly {
		pyexe := qpy.GetPython(py == "py3")
		if Fstdout == "" || Ftransported {
			fmt.Printf("%s", pyexe)
			return nil
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		fmt.Fprintf(f, "%s", pyexe)
		return nil
	}

	dirname := filepath.Dir(pyscript)
	basename := filepath.Dir(pyscript)
	d := new(qclient.Dir)
	d.Dir = dirname
	locfil := d.Get(basename)
	version := ""
	if locfil != nil {
		version = locfil.Release
	}
	project := ""
	if locfil != nil {
		project = locfil.Project
	}
	qpath := ""
	if locfil != nil {
		qpath = locfil.QPath
	}

	extra := []string{
		"VERSION__ = '" + version + "'",
		"PROJECT__ = '" + project + "'",
		"QPATH__ = '" + qpath + "'",
	}
	args = args[1:]

	sout, serr := qpy.Run(pyscript, py3, args, extra, Fcwd)

	if !stderrHidden && serr != "" {
		os.Stderr.WriteString(serr)
	}
	if !stdoutHidden && sout != "" {
		os.Stdout.WriteString(sout)
	}
	return nil
}
