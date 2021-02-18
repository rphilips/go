package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	qsync "brocade.be/qtechng/lib/sync"
)

// Fnextversion indicates the next version
var Fnextversion string

var versionCloseCmd = &cobra.Command{
	Use:     "close",
	Short:   "Closes a release",
	Long:    `A release is closed and the repository is copied to the appropriate number`,
	Args:    cobra.NoArgs,
	Example: "qtechng version close --nextversion=5.30",
	RunE:    versionClose,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"with-qtechtype": "B",
		"fill-version":   "yes",
	},
}

func init() {
	versionCloseCmd.PersistentFlags().StringVar(&Fnextversion, "nextversion", "", "next to develop version")
	versionCmd.AddCommand(versionCloseCmd)
}

func versionClose(cmd *cobra.Command, args []string) error {
	nextversion := Fnextversion
	if Fnextversion == "" || Fnextversion == "0.00" {
		err := fmt.Errorf("Next version `%s` is invalid", nextversion)
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	current := qregistry.Registry["brocade-release"]
	br := strings.TrimRight(current, " -_betaBETA")
	Fnextversion = strings.TrimRight(Fnextversion, " -_betaBETA")

	Fnextversion = qserver.Canon(Fnextversion)

	if br == Fnextversion {
		err := fmt.Errorf("Version `%s` is already closed", br)
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	_, err := qserver.Release{}.New(Fnextversion, true)
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	lowest := qserver.Lowest(Fnextversion, br)
	if lowest == Fnextversion {
		err = &qerror.QError{
			Ref: []string{"close.version.lowest"},
			Msg: []string{"The version of the new release `" + nextversion + "` should be higher than `" + br + "`"},
		}
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	_, _, err = qsync.Sync("0.00", br, true)

	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}

	err = qregistry.SetRegistry("brocade-release", Fnextversion)
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	err = qregistry.SetRegistry("brocade-release-say", Fnextversion+"beta")
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	x := qregistry.Registry["brocade-releases"]
	if x != "" {
		x += " "
	}
	if !strings.Contains(" "+x+" ", " "+br+" ") {
		err = qregistry.SetRegistry("brocade-releases", x+br)
	}
	if err != nil {
		Fmsg = qerror.ShowResult(Fmsg, Fjq, err)
		return nil
	}
	return nil
}
