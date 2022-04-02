package cmd

import (
	"path/filepath"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List qtechng files",
	Long:    `Lists local qtechng files and their properties` + Mfiles,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng file list application/bcawedit.m install.py`,
	RunE:    fileList,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

//Fonlychanged flag to indicate only changed files
var Fonlychanged bool

func init() {
	fileListCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileListCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileListCmd.Flags().BoolVar(&Fonlychanged, "changed", false, "Consider only modified files")
	fileListCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileListCmd)
}

func fileList(cmd *cobra.Command, args []string) error {
	plocfils, errlist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, Fonlychanged, Finlist, Fnotinlist, nil)
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

	list := make([]string, 0)
	result := make([]adder, 0)
	for _, locfil := range plocfils {
		changed := locfil.Changed(locfil.Place)
		rel, _ := filepath.Rel(Fcwd, locfil.Place)
		result = append(result, adder{rel, changed, locfil.Release, locfil.QPath, locfil.Place, qutil.FileURL(locfil.Place, locfil.QPath, -1), locfil.Time, locfil.Digest, locfil.Cu, locfil.Mu, locfil.Ct, locfil.Mt})
		if Flist != "" {
			list = append(list, locfil.QPath)
		}
	}
	Fmsg = qreport.Report(result, errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	if len(list) != 0 {
		qutil.EditList(Flist, false, list)
	}

	return nil
}
