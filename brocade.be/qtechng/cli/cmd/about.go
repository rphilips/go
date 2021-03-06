package cmd

import (
	"encoding/hex"
	"os"
	"os/user"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Information about `qtechng`",
	Long: `
Version and builttime information about qtechng.
If arguments are given, they are shown in 'hexified' format.`,
	Args:    cobra.ArbitraryArgs,
	Example: "  qtechng about\n  qtechng about --remote",
	RunE:    about,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
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
	if len(args) != 0 {
		for _, arg := range args {
			msg["hexified "+arg] = hex.EncodeToString([]byte(arg))
		}
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
