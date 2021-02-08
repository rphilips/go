package cmd

import (
	"fmt"
	"os/exec"
	"path"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var registryCmd = &cobra.Command{
	Use:     "registry",
	Short:   "registry functions",
	Long:    `All kinds of actions on the registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng registry",
}

type regitem struct {
	name      string
	ask       bool
	qtechtype string
	test      func(string) string
	nature    string
	deffunc   func() string
	doc       string
}

func testexe(value string) string {
	if value == "" {
		return "should not be empty"
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(value, ".exe") {
		return "should end on `.exe`"
	}
	basename := path.Base(value)
	if value != basename {
		return "give only the basename"
	}
	_, err := exec.LookPath(value)
	if err != nil {
		return fmt.Sprintf("cannot find `%s` in PATH", value)
	}
	return ""
}

var regmap2 = []regitem{
	{
		name:      "qtechng-exe",
		ask:       true,
		qtechtype: "BPW",
		test:      testexe,
		nature:    "exe",
		deffunc: func() string {
			if runtime.GOOS == "windows" {
				return "qtechng.exe"
			}
			return "qtechng"
		},
		doc: "Name of the basename of the `qtechng` binary. Take care is installed in the :envvar:`PATH`.\n\nExamples are:\n\n    - `qtechng` (on Linux and OSX)`\n    - `qtechng.exe` (on Windows)",
	},
	{
		name:      "qtechng-copy-exe",
		ask:       false,
		qtechtype: "BP",
		test:      func(string) string { return "" },
		nature:    "json",
		deffunc: func() string {
			return "[\"rsync\", \"-ai\", \"--delete\", \"--exclude=source/.hg\",  \"--exclude=source/.git\", \"/library/repository/{versionsource}/\", \"/library/repository/{versiontarget}\"]"
		},
		doc: "Action to execute for copying data from one version (`versionsource`) to the other (`versiontarget`).\nExecutable and arguments are in a `JSON` array. The literal names `versionsource` and `versiontarget` are placeholders in teh arguments.\nThey are between `{` and `}`",
	},
	{
		name:      "qtechng-sync-exe",
		ask:       false,
		qtechtype: "P",
		test:      func(string) string { return "" },
		nature:    "json",
		deffunc: func() string {
			return "[\"rsync\", \"-azi\", \"--delete\", \"--exclude=source/.hg\",  \"--exclude=source/.git\", \"root@dev.anet.be:/library/repository/{versionsource}/\", \"/library/repository/{versiontarget}\"]"
		},
		doc: "Action to execute on a productionserver for syncing the Brocade software from dev.anet.be.\nExecutable and arguments are in a `JSON` array. The literal names `versionsource` and `versiontarget` are placeholders in teh arguments.\nThey are between `{` and `}`",
	},
}

// map[string][4]string{
// 	"qtechng-copy-exe":            {"BP", "", ""},
// 	"qtechng-sync-exe":            {"BP", "", ""},
// 	"qtechng-diff":                {"BW", "", "exe"},
// 	"qtechng-support-dir":         {"W", "", "dir"},
// 	"qtechng-editor-exe":          {"W", "", "exe"},
// 	"qtechng-git-enable":          {"B", "[01]", ""},
// 	"qtechng-git-backup":          {"B", "", ""},
// 	"qtechng-log":                 {"W", "", "file"},
// 	"qtechng-max-parallel":        {"BWP", "", ""},
// 	"qtechng-repository-dir":      {"BP", "", "dir"},
// 	"qtechng-server":              {"WP", "", ""},
// 	"qtechng-test":                {"BWP", "test-entry", ""},
// 	"qtechng-type":                {"", "[BPW]+", ""},
// 	"qtechng-user":                {"WBP", "[^ ].*", ""},
// 	"qtechng-unique-ext":          {"B", "[^ ].*", ""},
// 	"qtechng-version":             {"WB", "[0-9]+\\.[0-9]+", ""},
// 	"qtechng-workstation-basedir": {"W", "[^ ].*", "dir"},
// 	"qtechng-block-qtechng":       {"B", "[01]", ""},
// 	"qtechng-block-doc":           {"BP", "[01]", ""},
// 	"qtechng-merge-exe":           {"W", "[^ ].*", "exe"},
// 	"brocade-release":             {"BP", "[0-9]+\\.[0-9][0-9][a-zA-Z]*"},
// 	"system-name":                 {"", "[^ ].*", ""},
// 	"system-group":                {"", "[^ ].*", ""},
// 	"system-roles":                {"BP", "[^ ].*", ""},
// 	"private-instname":            {"", "", ""},
// 	"private-role":                {"", "", ""},
// 	"fs-owner-qtechng":            {"BWP", "", ""},
// 	"lock-dir":                    {"BP", "[^ ].*", "dir"},
// 	"scratch-dir":                 {"WBP", "[^ ].*", "dir"},
// 	"m-os-type":                   {"BP", "[^ ].*", ""},
// 	"gtm-rou-dir":                 {"BP", "[^ ].*", "dir"},
// 	"m-import-auto-exe":           {"BP", "[^ ].*", "exe"},
// 	"os":                          {"BP", "[^ ].*", ""},
// 	"m-clib":                      {"BP", "[^ ].*", ""},
// 	"web-base-url":                {"BP", "[^ ].*", ""},
// }

func init() {
	registryCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	rootCmd.AddCommand(registryCmd)
}
