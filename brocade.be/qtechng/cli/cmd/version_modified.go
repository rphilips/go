package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var versionModifiedCmd = &cobra.Command{
	Use:     "modified",
	Short:   "Show modification since a time",
	Long:    `This command shows the modified files since a timestamp`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng version modified 0.00`,
	RunE:    versionModified,
	Annotations: map[string]string{
		"with-qtechtype": "BPW",
	},
}

var Fafter string

func init() {
	versionCmd.AddCommand(versionModifiedCmd)
	versionModifiedCmd.Flags().StringVar(&Fafter, "after", "", "Give timestamp")
}

func versionModified(cmd *cobra.Command, args []string) error {
	qtechType := qregistry.Registry["qtechng-type"]
	version := ""
	if len(args) == 0 {
		if strings.ContainsRune(qtechType, 'P') {
			version = qregistry.Registry["brocade-release"]
		} else {
			version = "0.00"
		}
	} else {
		version = args[0]
	}
	if !strings.ContainsAny(qtechType, "BW") && version != qregistry.Registry["brocade-release"] {
		err := &qerror.QError{
			Ref: []string{"modified.version"},
			Msg: []string{"The version  should be `" + qregistry.Registry["brocade-release"] + "`"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	release, err := qserver.Release{}.New(version, true)
	if strings.ContainsAny(QtechType, "PB") {
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"modified.new"},
				Version: version,
				Msg:     []string{"Cannot instantiate version"},
			}
			err = qerror.QErrorTune(err, e)
			return err
		}

		if Fafter == "" && !strings.ContainsAny(qtechType, "BW") {
			tsf, _ := release.FS("/").RealPath("/admin/sync.json")
			tsb, err := qfs.Fetch(tsf)
			Fafter = "2020-01-01T00:00:00"
			if err == nil {
				data := make(map[string]string)
				err := json.Unmarshal(tsb, &data)
				if err == nil && data["timestamp"] != "" {
					Fafter = data["timestamp"]
				}
			}
		}
	}
	tim, err := qutil.TimeParse(Fafter)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	if strings.ContainsRune(QtechType, 'B') {
		after, _ := qutil.TimeParse(Fafter)
		result, err := release.Modifications(after, "")
		Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	Fafter = tim.Format(time.RFC3339Nano)

	argums := []string{
		"ssh",
		"qtechng",
		"version",
		"modified",
		version,
		"--after=" + Fafter,
	}
	stdout, stderr, err := qutil.QtechNG(argums, Fjq, Fyaml, Fcwd)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if strings.TrimSpace(stderr) != "" {
		Fmsg = qreport.Report(nil, errors.New(stderr), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	fmt.Println(stdout)

	return nil
}
