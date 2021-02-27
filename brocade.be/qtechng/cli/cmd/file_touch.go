package cmd

import (
	"os"
	"path/filepath"
	"time"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
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
	plocfils, errlist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern)

	h := time.Now().Local()
	t := h.Format(time.RFC3339)
	type adder struct {
		Name    string `json:"arg"`
		Release string `json:"version"`
		Qpath   string `json:"qpath"`
		Time    string `json:"time"`
	}

	result := make([]adder, 0)
	for _, file := range plocfils {
		place := file.Place
		et := os.Chtimes(file.Place, h, h)
		if et == nil {
			rel, _ := filepath.Rel(Fcwd, place)
			result = append(result, adder{rel, file.Release, file.QPath, t})
		}
	}
	Fmsg = qerror.ShowResult(result, Fjq, errlist)
	return nil
}
