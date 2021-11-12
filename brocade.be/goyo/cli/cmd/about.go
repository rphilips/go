package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"time"

	qyottadb "brocade.be/goyo/lib/yottadb"
	"github.com/spf13/cobra"
)

var aboutCmd = &cobra.Command{
	Use:     "about",
	Short:   "Information about `goyo`",
	Long:    `Version and build time information about the goyo executable.`,
	Args:    cobra.NoArgs,
	Example: `goyo about`,
	RunE:    about,
}

func init() {
	rootCmd.AddCommand(aboutCmd)
}

func about(cmd *cobra.Command, args []string) error {
	msg := map[string]string{"BuildTime": BuildTime, "BuildHost": BuildHost, "BuildWith": GoVersion}
	host, e := os.Hostname()

	if e == nil {
		msg["uname"] = host
	}
	user, err := user.Current()
	if err == nil {
		msg["user.name"] = user.Name
		msg["user.username"] = user.Username
	}

	err = qyottadb.Set("/zgoya/hello", "Hello World")
	if err != nil {
		msg["error on set"] = err.Error()
	}
	b, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))

	return nil
}

func AboutText() string {

	msg := map[string]string{"BuildTime": BuildTime, "BuildHost": BuildHost, "BuildWith": GoVersion}
	host, e := os.Hostname()

	if e == nil {
		msg["uname"] = host
	}
	user, err := user.Current()
	if err == nil {
		msg["user.name"] = user.Name
		msg["user.username"] = user.Username
	}

	h := time.Now()
	now := h.Format(time.RFC3339)
	qyottadb.Set("/zgoya/last/"+user.Name, now)
	dbnow, _ := qyottadb.G("/zgoya/last/"+user.Name, false)
	if now == dbnow {
		msg["status"] = "Connected successfully to YottaDB!"
	} else {
		msg["status"] = "Failed to connectto YottaDB!"
	}
	msg["time"] = now
	b, _ := json.MarshalIndent(msg, "", "  ")
	return string(b)
}
