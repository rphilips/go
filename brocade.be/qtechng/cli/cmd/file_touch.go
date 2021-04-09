package cmd

import (
	"path/filepath"
	"time"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileTouchCmd = &cobra.Command{
	Use:     "touch",
	Short:   "Touches QtechNG files",
	Long:    `Command die de mtime/atime van een bestand aanpast.`,
	Args:    cobra.MinimumNArgs(0),
	Example: `  qtechng file touch application/bcawedit.m install.py cwd=../catalografie`,
	RunE:    fileTouch,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

func init() {
	fileTouchCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileTouchCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileTouchCmd.Flags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileTouchCmd)
}

func fileTouch(cmd *cobra.Command, args []string) error {
	plocfils, errlist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false)

	h := time.Now().Local()
	t := h.Format(time.RFC3339)
	type adder struct {
		Name    string `json:"arg"`
		Release string `json:"version"`
		QPath   string `json:"qpath"`
		Place   string `json:"file"`
		Url     string `json:"fileurl"`
		Time    string `json:"time"`
	}

	result := make([]adder, 0)
	tail := make([]byte, 0)

	errslice := make([]error, 0)
	if errlist != nil {
		errslice = append(errslice, errlist)
	}

	for _, file := range plocfils {
		place := file.Place
		et := qfs.Append(place, tail)
		if et == nil {
			rel, _ := filepath.Rel(Fcwd, place)
			result = append(result, adder{rel, file.Release, file.QPath, place, qutil.FileURL(place, -1), t})
		} else {
			errslice = append(errslice, et)
		}
	}
	Fmsg = qreport.Report(result, errslice, Fjq, Fyaml)
	return nil
}
