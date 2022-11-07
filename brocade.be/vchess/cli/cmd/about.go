package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"

	vregistry "brocade.be/vchess/lib/registry"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Information about `vchess`",
	Long:  `Version and build time information about the vchess executable`,

	Args:    cobra.NoArgs,
	Example: `vchess about`,
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

	_, ok := vregistry.Registry["error"]
	if ok {
		return fmt.Errorf(vregistry.Registry["error"].(string))
	}
	msg["REGISTRY"] = os.Getenv("MY_REGISTRY")

	b, _ := json.MarshalIndent(msg, "", "    ")
	Fmsg = string(b)
	return nil
}
