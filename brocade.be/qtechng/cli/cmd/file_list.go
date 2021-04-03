package cmd

import (
	"path/filepath"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists QtechNG files",
	Long:    `Command `,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng file list application/bcawedit.m install.py`,
	RunE:    fileList,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

//Fonlychanged flag to indicate only changed files
var Fonlychanged bool

func init() {
	fileListCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileListCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileListCmd.Flags().BoolVar(&Fonlychanged, "changed", false, "Consider only modified files")
	fileListCmd.Flags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileListCmd)
}

func fileList(cmd *cobra.Command, args []string) error {
	plocfils, errlist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, Fonlychanged)
	type adder struct {
		Name    string `json:"arg"`
		Changed bool   `json:"changed"`
		Release string `json:"version"`
		QPath   string `json:"qpath"`
		Path    string `json:"file"`
		URL     string `json:"fileurl"`
		Time    string `json:"time"`
		Digest  string `json:"digest"`
		Cu      string `json:"cu"`
		Mu      string `json:"mu"`
		Ct      string `json:"ct"`
		Mt      string `json:"mt"`
	}

	result := make([]adder, 0)
	for _, locfil := range plocfils {
		changed := locfil.Changed(locfil.Place)
		rel, _ := filepath.Rel(Fcwd, locfil.Place)
		result = append(result, adder{rel, changed, locfil.Release, locfil.QPath, locfil.Place, qutil.FileURL(locfil.Place, -1), locfil.Time, locfil.Digest, locfil.Cu, locfil.Mu, locfil.Ct, locfil.Mt})
	}
	Fmsg = qerror.ShowResult(result, Fjq, errlist, Fyaml)
	return nil
}
