package cmd

import (
	"encoding/json"
	"os"
	"os/user"

	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Information about `pblad`",
	Long:  `Version and build time information about the qtechng executable.`,

	Args:    cobra.NoArgs,
	Example: `pblad about`,
	RunE:    about,
}

func init() {

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
	msg["REGISTRY"] = os.Getenv("MY_REGISTRY")
	b, _ := json.MarshalIndent(msg, "", "    ")
	Fmsg = string(b)
	return nil
}
