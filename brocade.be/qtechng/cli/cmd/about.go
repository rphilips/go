package cmd

import (
	"errors"
	"os"
	"os/user"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Information about `qtechng`",
	Long: `Version and build time information about the qtechng executable.

The '--remote' flag can be used to give information about the 'qtechng' executable on the development server.`,
	Args: cobra.NoArgs,
	Example: `qtechng about
qtechng about --remote`,
	RunE:   about,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"complete":       "end",
	},
}

func init() {

	aboutCmd.Flags().BoolVar(&Fremote, "remote", false, "Execute on the remote server")
	rootCmd.AddCommand(aboutCmd)
}

func about(cmd *cobra.Command, args []string) error {
	if qregistry.Registry["error"] != "" {
		Fmsg = qreport.Report(nil, errors.New(qregistry.Registry["error"]), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	msg := map[string]string{"!BuildTime": BuildTime, "!BuildHost": BuildHost, "!BuildWith": GoVersion}
	host, e := os.Hostname()

	if e == nil {
		msg["!!uname"] = host
	}
	user, err := user.Current()
	if err == nil {
		msg["!!user.name"] = user.Name
		msg["!!user.username"] = user.Username
	}
	msg["BROCADE_REGISTRY"] = os.Getenv("BROCADE_REGISTRY")
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
