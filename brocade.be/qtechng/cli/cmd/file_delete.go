package cmd

import (
	"path/filepath"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete qtechng status of files",
	Long: `Removes the status of one or more local files as qtechng files.

Remember:
    - This action has no effect on the repository
    - Use the '--unlink' flag to also remove the files from the filesystem!


` + Mfiles,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng file delete application/bcawedit.m install.py cwd=../workspace
qtechng file delete test.rst --unlink`,
	RunE: fileDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

var Funlink bool

func init() {
	fileDeleteCmd.Flags().BoolVar(&Funlink, "unlink", false, "Remove from filesystem")
	fileDeleteCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileDeleteCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileDeleteCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileDeleteCmd)
}

func fileDelete(cmd *cobra.Command, args []string) error {
	type deleter struct {
		Name    string `json:"arg"`
		Release string `json:"version"`
		QPath   string `json:"qpath"`
	}

	done := make(map[string]bool)

	plocfils, err := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false, Finlist, Fnotinlist, nil)
	direxists := make(map[string][]qclient.LocalFile)

	result := make([]deleter, 0)
	errorlist := make([]error, 0)
	if err != nil {
		errorlist = append(errorlist, err)
	}
	list := make([]string, 0)
	for _, plocfil := range plocfils {
		place := plocfil.Place
		if done[place] {
			continue
		}
		done[place] = true
		if Flist != "" {
			list = append(list, plocfil.QPath)
		}

		dir := filepath.Dir(place)

		_, ok := direxists[dir]

		if !ok {
			direxists[dir] = make([]qclient.LocalFile, 0)
		}
		direxists[dir] = append(direxists[dir], *plocfil)

	}

	for dir, locfils := range direxists {
		dir := qclient.Dir{
			Dir: dir,
		}
		dir.Del(locfils...)
		for _, locfil := range locfils {
			result = append(result, deleter{
				Name:    locfil.Place,
				Release: locfil.Release,
				QPath:   locfil.QPath,
			})
			if Funlink {
				qfs.Rmpath(locfil.Place)
			}
		}
	}

	Fmsg = qreport.Report(result, errorlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	if len(list) != 0 {
		qutil.EditList(Flist, false, list)
	}
	return nil
}
