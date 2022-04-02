package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"

	qfs "brocade.be/base/fs"
	qpy "brocade.be/base/python"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

//Ftell tells what kind of informatiom has to be returned
var Ftell = ""
var fileTellCmd = &cobra.Command{
	Use:   "tell",
	Short: "Give information about files",
	Long: `Gives information about files (to be used in shell scripts)

The '--tell' flag specifies which information has to be displayed.
(without this flag, all information is given)

'--tell' can have the following values:

    - ext: file extension
	- basename: basename of the file
	- dirname: directory name
	- abspath: complete file specification
	- version: repository version
	- project: project
	- qpath: qpath
	- qdir: repository directory
	- python: Python executable
	- relpath: relative path of qpath versus project
	- fileurl: URL with file scheme
	- vcurl: URL in version control
	- changed: true/false

Note: this information is retrieved locally and can be outdated.

	` + Mfiles,

	Example: `  qtechng file tell bcawedit.m --cwd=../workspace --ext
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=dirname
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=basename
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=project
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=ext
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=qpath
	  qtechng file tell bcawedit.m --cwd=../workspace --tell=version
	  qtechng file tell bcawedit.m --cwd=../workspace
	`,
	RunE: fileTell,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	fileTellCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileTellCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileTellCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileTellCmd.Flags().StringVar(&Ftell, "tell", "", "abspath/relpath/ext/dirname/basename/version/project/qpath/python")
	fileCmd.AddCommand(fileTellCmd)
}

func fileTell(cmd *cobra.Command, args []string) error {

	plocfils, _ := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false, Finlist, Fnotinlist, nil)

	if len(plocfils) == 0 {
		err := qerror.QError{
			Ref: []string{"tell.args.0"},
			Msg: []string{"Cannot find a file with these specifications"},
		}
		return err
	}

	if len(plocfils) != 1 {
		err := qerror.QError{
			Ref: []string{"tell.args.2"},
			Msg: []string{"Too many files found: need exactly 1"},
		}
		return err
	}

	locfil := plocfils[0]
	if locfil == nil {
		return nil
	}

	fname, _ := qfs.AbsPath(locfil.Place)
	relpath, _ := filepath.Rel(Fcwd, fname)
	dirname := filepath.Dir(fname)
	basename := filepath.Base(fname)
	result := make(map[string]string)
	result["ext"] = path.Ext(args[0])
	result["basename"] = basename
	result["dirname"] = dirname
	result["abspath"] = fname
	result["version"] = ""
	result["project"] = ""
	result["qpath"] = ""
	result["qdir"] = ""
	result["python"] = ""
	result["relpath"] = relpath
	result["fileurl"] = qutil.FileURL(fname, "", -1)
	result["changed"] = "false"
	if locfil.Changed(locfil.Place) {
		result["changed"] = "true"
	}

	py := qutil.GetPy(fname, filepath.Dir(fname))
	if py != "" {
		pyexe := qpy.GetPython(py == "py3")
		result["python"] = pyexe
	}

	result["version"] = locfil.Release
	result["project"] = locfil.Project
	result["qpath"] = locfil.QPath
	result["vcurl"] = qutil.VCURL(locfil.QPath)
	qdir, _ := qutil.QPartition(locfil.QPath)
	result["qdir"] = qdir

	tell, ok := result[Ftell]

	if ok {
		if Fstdout == "" || Ftransported {
			fmt.Print(tell)
			return nil
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		fmt.Fprint(w, tell)
		err = w.Flush()
		return err
	}
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
