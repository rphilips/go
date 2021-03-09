package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
)

var versionDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes a release",
	Long:    `A release is deleted from the repository`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version delete 5.30",
	RunE:    versionDelete,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	versionDeleteCmd.PersistentFlags().BoolVar(&Fforce, "force", false, "with force")
	versionCmd.AddCommand(versionDeleteCmd)
}

func versionDelete(cmd *cobra.Command, args []string) error {
	if !Fforce && strings.Contains(QtechType, "B") {
		err := fmt.Errorf("On a development server, this command can only be used with `force`")
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	version := args[0]
	version = qserver.Canon(version)

	if version == "0.00" || version == "" {
		err := fmt.Errorf("Version `0.00` cannot be deleted")
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	br := qregistry.Registry["brocade-release"]
	br = strings.TrimRight(br, " -_betaBETA")

	if strings.Contains(QtechType, "B") && br == version {
		err := fmt.Errorf("Current version `" + br + "` cannot be deleted")
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		Fmsg = qerror.ShowResult("", Fjq, err)
		return nil
	}

	ok, _ := release.Exists("")
	if !ok {
		err = fmt.Errorf("Version `%s` does NOT exist", release.String())
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	err = qfs.Rmpath(release.Root())
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	return nil
}