package cmd

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qmeta "brocade.be/qtechng/lib/meta"
	qproject "brocade.be/qtechng/lib/project"
	qreport "brocade.be/qtechng/lib/report"
)

var projectNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Creates a new project",
	Long: `Command creates a new project on the development server.
	With the treeprefix not empty, starting from the current working directory, all projects 
	are installed. The name of the project is determied by the relative path of the directories 
	and prefixed with the treeprefix.
	`,
	Args:    cobra.MinimumNArgs(0),
	Example: "qtechng project new /stdlib/template\nqtechng project new /stdlib/template  --version=5.10",
	RunE:    projectNew,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"always-remote":  "yes",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

//Ftreeprefix defines the prefix which has to prepended to the directories
var Ftreeprefix string

func init() {
	projectNewCmd.Flags().StringVar(&Ftreeprefix, "treeprefix", "", "Find projects starting with cwd")
	projectCmd.AddCommand(projectNewCmd)
}

func projectNew(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && strings.HasPrefix(Ftreeprefix, "/") {
		Ftreeprefix = strings.TrimRight(Ftreeprefix, "/") + "/"
		argums, err := qfs.Find(Fcwd, []string{"brocade.json"}, true, true, false)
		if err != nil {
			return err
		}
		for _, arg := range argums {
			dirname := filepath.Dir(arg)
			rel, _ := filepath.Rel(Fcwd, dirname)
			rel = filepath.Clean(rel)
			rel = filepath.ToSlash(rel)
			rel = strings.TrimPrefix(rel, "./")
			rel = strings.TrimRight(rel, "/")
			rel = strings.TrimLeft(rel, "/")
			args = append(args, Ftreeprefix+rel)
		}
	}
	meta := qmeta.Meta{
		Mu: FUID,
	}
	result, errs := qproject.InitList(Fversion, args, func(a string) qmeta.Meta { return meta })

	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
