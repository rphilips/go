package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:     "system",
	Short:   "System information",
	Long:    `Command allows for  configuration/testing of setup`,
	Args:    cobra.NoArgs,
	Example: "qtechng system",
}

func init() {
	systemCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	rootCmd.AddCommand(systemCmd)
}

func regMap() map[string][3]string {
	regmap := map[string][3]string{
		"qtechng-binary":              {"", "", ""},
		"qtechng-copy-exe":            {"BP", "", ""},
		"qtechng-sync-exe":            {"BP", "", ""},
		"qtechng-diff":                {"BW", "", "exe"},
		"qtechng-support-dir":         {"W", "", "dir"},
		"qtechng-editor-exe":          {"W", "", "exe"},
		"qtechng-git-enable":          {"B", "[01]", ""},
		"qtechng-git-backup":          {"B", "", ""},
		"qtechng-log":                 {"W", "", "file"},
		"qtechng-max-parallel":        {"BWP", "", ""},
		"qtechng-repository-dir":      {"BP", "", "dir"},
		"qtechng-server":              {"WP", "", ""},
		"qtechng-test":                {"BWP", "test-entry", ""},
		"qtechng-type":                {"", "[BPW]+", ""},
		"qtechng-user":                {"WBP", "[^ ].*", ""},
		"qtechng-unique-ext":          {"B", "[^ ].*", ""},
		"qtechng-version":             {"WB", "[0-9]+\\.[0-9]+", ""},
		"qtechng-workstation-basedir": {"W", "[^ ].*", "dir"},
		"qtechng-block-qtechng":       {"B", "[01]", ""},
		"qtechng-block-doc":           {"BP", "[01]", ""},
		"qtechng-merge-exe":           {"W", "[^ ].*", "exe"},
		"brocade-release":             {"BP", "[0-9]+\\.[0-9][0-9][a-zA-Z]*"},
		"system-name":                 {"", "[^ ].*", ""},
		"system-group":                {"", "[^ ].*", ""},
		"system-roles":                {"BP", "[^ ].*", ""},
		"private-instname":            {"", "", ""},
		"private-role":                {"", "", ""},
		"fs-owner-qtechng":            {"BWP", "", ""},
		"lock-dir":                    {"BP", "[^ ].*", "dir"},
		"scratch-dir":                 {"WBP", "[^ ].*", "dir"},
		"m-os-type":                   {"BP", "[^ ].*", ""},
		"gtm-rou-dir":                 {"BP", "[^ ].*", "dir"},
		"m-import-auto-exe":           {"BP", "[^ ].*", "exe"},
		"os":                          {"BP", "[^ ].*", ""},
		"m-clib":                      {"BP", "[^ ].*", ""},
		"web-base-url":                {"BP", "[^ ].*", ""},
	}
	for key, value := range regmap {
		if value[0] != "" && !strings.Contains(value[0], QtechType) {
			delete(regmap, key)
		}
	}
	return regmap
}
