package cmd

import (
	"path/filepath"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fileDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes QtechNG files",
	Long:    `Deletes the status of a file as a QtechNG file`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng file delete application/bcawedit.m install.py cwd=../catalografie`,
	RunE:    fileDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

func init() {
	fileDeleteCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileDeleteCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileDeleteCmd.Flags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileDeleteCmd)
}

func fileDelete(cmd *cobra.Command, args []string) error {
	type deleter struct {
		Name    string `json:"arg"`
		Release string `json:"version"`
		QPath   string `json:"qpath"`
	}

	done := make(map[string]bool)
	if Fversion == "" {
		err := &qerror.QError{
			Ref:  []string{"file.delete.version"},
			Type: "Error",
			Msg:  []string{"Do not know how to deduce version"},
		}
		Fmsg = qreport.Report("", err, Fjq, Fyaml)
		return nil
	}
	if Fcwd == "" {
		err := &qerror.QError{
			Ref:  []string{"file.delete.cwd"},
			Type: "Error",
			Msg:  []string{"Do not know where to find the files"},
		}
		Fmsg = qreport.Report("", err, Fjq, Fyaml)
		return nil
	}

	plocfils, errlist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false)
	direxists := make(map[string][]qclient.LocalFile)

	result := make([]deleter, 0)
	errorlist := make([]error, 0)
	for _, plocfil := range plocfils {
		place := plocfil.Place
		if done[place] {
			continue
		}
		done[place] = true

		if qfs.IsDir(place) {
			err := &qerror.QError{
				Ref:  []string{"file.delete.dir"},
				Type: "Error",
				Msg:  []string{"Cannot delete a directory: `" + place + "`"},
			}
			errorlist = append(errorlist, err)
			continue
		}

		dir := filepath.Dir(place)

		_, ok := direxists[dir]

		if !ok {
			direxists[dir] = make([]qclient.LocalFile, 0)
			direxists[dir] = append(direxists[dir], *plocfil)
		}

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
		}
	}
	Fmsg = qreport.Report(result, errlist, Fjq, Fyaml)
	return nil
}
