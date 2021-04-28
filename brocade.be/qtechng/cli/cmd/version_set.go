package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var versionSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Sets version number",
	Long:    `The required version numbers are set`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version set 5.40",
	RunE:    versionSet,
	Annotations: map[string]string{
		"with-qtechtype": "P",
	},
}

func init() {
	versionCmd.AddCommand(versionSetCmd)
}

func versionSet(cmd *cobra.Command, args []string) error {
	version := args[0]
	version = qserver.Canon(version)

	if strings.Contains(QtechType, "B") {
		err := fmt.Errorf("On a development server, this command cannot be used")
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}

	if version == "0.00" || version == "" {
		err := fmt.Errorf("Version `0.00` cannot be set")
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}

	br := qregistry.Registry["brocade-release"]
	br = strings.TrimRight(br, " -_betaBETA")

	lowest := qutil.LowestVersion(version, br)
	if lowest == version {
		err := &qerror.QError{
			Ref: []string{"set.version.lowest"},
			Msg: []string{"The version of the new release `" + version + "` should be higher than `" + br + "`"},
		}
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}

	ok, _ := release.Exists("")
	if !ok {
		err = fmt.Errorf("Version `%s` does NOT exist", release.String())
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}

	err = qregistry.SetRegistry("brocade-release", version)
	if err != nil {
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}
	err = qregistry.SetRegistry("brocade-release-say", version)
	if err != nil {
		Fmsg = qreport.Report(Fmsg, err, Fjq, Fyaml)
		return nil
	}
	return nil
}
