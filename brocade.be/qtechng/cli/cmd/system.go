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
		"qtechng-user":                {"W", "[^ ].*", ""},
		"qtechng-unique-ext":          {"B", "[^ ].*", ""},
		"qtechng-server":              {"W", "[^ ]+", ""},
		"qtechng-root-dir":            {"BP", "[^ ].*", "dir"},
		"qtechng-repository-dir":      {"BP", "[^ ].*", "dir"},
		"qtechng-version":             {"", "[0-9]+\\.[0-9][0-9]", ""},
		"qtechng-hg-backup":           {"B", "[^ ].*"},
		"qtechng-workstation-basedir": {"W", "[^ ].*", "dir"},
		"qtechng-support-project":     {"W", "[^ ].*", ""},
		"qtechng-test-dir":            {"W", "[^ ].*", "dir"},
		"qtechng-block-qtechng":       {"B", "[01]", ""},
		"qtechng-block-doc":           {"BP", "[01]", ""},
		"qtechng-releases":            {"W", "[^ ].*", ""},
		"qtechng-editor-exe":          {"W", "[^ ].*", "exe"},
		"qtechng-merge-exe":           {"W", "[^ ].*", "exe"},
		"qtechng-type":                {"", "[BPW]", ""},
		"qtechng-hg-enable":           {"B", "[01]"},
		"brocade-release":             {"P", "[0-9]+\\.[0-9][0-9][a-zA-Z]*"},
		"system-name":                 {"", "[^ ].*", ""},
	}
	for key, value := range regmap {
		if value[0] != "" && !strings.Contains(value[0], QtechType) {
			delete(regmap, key)
		}
	}
	return regmap
}
