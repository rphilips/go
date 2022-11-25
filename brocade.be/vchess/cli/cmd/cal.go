package cmd

import (
	"os/exec"
	"strings"

	vregistry "brocade.be/vchess/lib/registry"
	vstructure "brocade.be/vchess/lib/structure"
	"github.com/spf13/cobra"
)

var calCmd = &cobra.Command{
	Use:   "cal",
	Short: "Information print `vchess`",
	Long:  `Version and build time printrmation print the vchess executable`,

	Args:    cobra.NoArgs,
	Example: `vchess call`,
	RunE:    cal,
}

func init() {
	rootCmd.AddCommand(calCmd)
}

func cal(cmd *cobra.Command, args []string) (err error) {

	season := new(vstructure.Season)
	season.Init(nil)

	_, err = season.Calendar()

	if err != nil {
		return
	}

	calfile := season.CalendarFile("html")
	aviewer := vregistry.Registry["viewer"].(map[string]any)["html"].([]any)
	viewer := make([]string, 0)

	for _, piece := range aviewer {
		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", calfile))
	}
	vcmd := exec.Command(viewer[0], viewer[1:]...)
	err = vcmd.Start()
	if err != nil {
		panic(err)
	}
	return nil
}
