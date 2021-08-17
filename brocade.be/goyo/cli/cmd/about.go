package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/spf13/cobra"
	"lang.yottadb.com/go/yottadb"
)

var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Information about `goyo`",
	Long: `Version and build time information about the goyo executable.
If arguments are given, they are shown in 'hexified' format.`,
	Args:    cobra.ArbitraryArgs,
	Example: `goyo about`,
	RunE:    about,
}

func init() {
	rootCmd.AddCommand(aboutCmd)
}

func about(cmd *cobra.Command, args []string) error {
	defer yottadb.Exit()
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
	if len(args) != 0 {
		for _, arg := range args {
			msg["hexified "+arg] = hex.EncodeToString([]byte(arg))
		}
	}

	err = yottadb.SetValE(yottadb.NOTTP, nil, "Aloha, galaxy!", "^zhello", []string{"goya"})
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
	err = yottadb.SetValE(yottadb.NOTTP, nil, now, "^goya", []string{"goya", user.Name, "last"})
	msg["time"] = now
	if err != nil {
		msg["error on set"] = err.Error()
	}
	b, _ := json.MarshalIndent(msg, "", "  ")
	return string(b)
}
