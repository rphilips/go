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
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

//Ftell tells what kind of informatiom has to be returned
var Ftell = ""
var fileTellCmd = &cobra.Command{
	Use:   "tell",
	Short: "Gives information about files",
	Long:  `Gives information about files (to be used in shell scripts)`,
	Example: `  qtechng file tell bcawedit.m --cwd=../catalografie --ext
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=dirname
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=basename
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=project
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=ext
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=qpath
	  qtechng file tell bcawedit.m --cwd=../catalografie --tell=version
	  qtechng file tell bcawedit.m --cwd=../catalografie
	`,
	RunE: fileTell,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	fileTellCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileTellCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileTellCmd.Flags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileTellCmd.Flags().StringVar(&Ftell, "tell", "", "abspath/relpath/ext/dirname/basename/version/project/qpath/python")
	fileCmd.AddCommand(fileTellCmd)
}

func fileTell(cmd *cobra.Command, args []string) error {

	plocfils, _ := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false)

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
			Msg: []string{"Toom many files found: need exactly 1"},
		}
		return err
	}

	locfil := plocfils[0]

	fname, _ := qfs.AbsPath(locfil.Place)
	relpath, _ := filepath.Rel(Fcwd, fname)
	dirname := path.Dir(fname)
	basename := path.Base(fname)
	result := make(map[string]string)
	result["ext"] = path.Ext(args[0])
	result["basename"] = basename
	result["dirname"] = dirname
	result["abspath"] = fname
	result["version"] = ""
	result["project"] = ""
	result["qpath"] = ""
	result["python"] = ""
	result["relpath"] = relpath
	result["fileurl"] = qutil.FileURL(fname, -1)

	py := qutil.GetPy(fname)
	if py != "" {
		pyexe := qpy.GetPython(py == "py3")
		result["python"] = pyexe
	}

	if locfil != nil {
		result["version"] = locfil.Release
		result["project"] = locfil.Project
		result["qpath"] = locfil.QPath

	}

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
	Fmsg = qerror.ShowResult(result, Fjq, nil)
	return nil
}
