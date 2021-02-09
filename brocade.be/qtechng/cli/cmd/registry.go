package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
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
	mode      string
	qtechtype string
	test      func(string) string
	nature    string
	deffunc   func() string
	doc       string
}

func testexe(value string) string {
	if strings.HasPrefix(value, "[") {
		args := make([]string, 0)
		err := json.Unmarshal([]byte(value), &args)
		if err != nil {
			return "illegal value"
		}
		if len(args) == 0 {
			return "empty list"
		}
		value = args[0]
	}
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

func isdir(value string) string {
	err := os.MkdirAll(value, 0755)
	if err == nil {
		return ""
	}
	return fmt.Sprintf("`%s` is not a directory or does not exist", value)
}

func dirname(dir string) string {
	dir = filepath.FromSlash(dir)
	if !path.IsAbs(dir) {
		home, _ := os.UserHomeDir()
		dir = path.Join(home, dir)
	}
	return dir
}

func isnum(value string) string {
	_, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Sprintf("`%s` is not a number", value)
	}
	return ""
}

func isbool(value string) string {
	if value != "" && value != "1" && value != "0" {
		return fmt.Sprintf("`%s` should be empty, `0` or `1`", value)
	}
	return ""
}

func isbwp(value string) string {
	x := strings.TrimLeft(value, "BWP")
	if x != "" {
		return fmt.Sprintf("`%s` should only contain `B`, `P` or `W`", value)
	}
	return ""
}

var regmapqtechng = []regitem{
	{
		name:      "qtechng-exe",
		mode:      "ask",
		qtechtype: "BPW",
		test:      testexe,
		nature:    "exe",
		deffunc: func() string {
			if runtime.GOOS == "windows" {
				return "qtechng.exe"
			}
			return "qtechng"
		},
		doc: "Basename of the `qtechng` binary. Take care is installed in the :envvar:`PATH`.\n\nExamples are:\n\n    - `qtechng` (on Linux and OSX)`\n    - `qtechng.exe` (on Windows)",
	},
	{
		name:      "qtechng-copy-exe",
		mode:      "skip",
		qtechtype: "BP",
		test:      func(string) string { return "" },
		nature:    "json",
		deffunc: func() string {
			return "[\"rsync\", \"-ai\", \"--delete\", \"--exclude=source/.hg\",  \"--exclude=source/.git\", \"/library/repository/{versionsource}/\", \"/library/repository/{versiontarget}\"]"
		},
		doc: "Action to execute for copying data from one version (`versionsource`) to the other (`versiontarget`).\nExecutable and arguments are in a `JSON` array. The literal names `versionsource` and `versiontarget` are placeholders in the arguments.\nThey are between `{` and `}`",
	},
	{
		name:      "qtechng-sync-exe",
		mode:      "skip",
		qtechtype: "P",
		test:      func(string) string { return "" },
		nature:    "json",
		deffunc: func() string {
			return "[\"rsync\", \"-azi\", \"--delete\", \"--exclude=source/.hg\",  \"--exclude=source/.git\", \"root@dev.anet.be:/library/repository/{versionsource}/\", \"/library/repository/{versiontarget}\"]"
		},
		doc: "Action to execute on a productionserver for syncing the Brocade software from dev.anet.be.\nExecutable and arguments are in a `JSON` array. The literal names `versionsource` and `versiontarget` are placeholders in the arguments.\nThey are between `{` and `}`",
	},
	{
		name:      "qtechng-diff-exe",
		mode:      "ask",
		qtechtype: "W",
		test:      testexe,
		nature:    "json",
		deffunc: func() string {
			switch runtime.GOOS {
			case "windows":
				return "WinMergeU.exe"
			case "linux":
				return "[\"meld\", \"{target}\", \"{source}\"]"
			case "darwin":
			}
			return ""
		},
		doc: "Local software to use for comparing two files.\nExecutable and arguments are in a `JSON` array. The literal names `target` and `source` are placeholders in the arguments.\nThey are between `{` and `}`",
	},
	{
		name:      "qtechng-support-dir",
		mode:      "ask",
		qtechtype: "WB",
		test:      isdir,
		nature:    "dir",
		deffunc: func() string {
			return dirname("brocade/support")
		},
		doc: "Local directory containing all kinds of data to help out with editing Brocade files",
	},
	{
		name:      "qtechng-editor-exe",
		mode:      "ask",
		qtechtype: "W",
		test:      testexe,
		nature:    "exe",
		deffunc: func() string {
			if runtime.GOOS == "windows" {
				return "code.exe"
			}
			return "code"
		},
		doc: "Basename of the prefered editor. Take care is installed in the :envvar:`PATH`.\n\nExamples are:\n\n    - `code` (on Linux and OSX)`\n    - `code.exe` (on Windows)",
	},
	{
		name:      "qtechng-git-enable",
		mode:      "skip",
		qtechtype: "B",
		test:      isbool,
		nature:    "bool",
		deffunc: func() string {
			return "1"
		},
		doc: "If `1`, version control with Git is enabled",
	},
	{
		name:      "qtechng-max-parallel",
		mode:      "skip",
		qtechtype: "BPW",
		test:      isnum,
		nature:    "number",
		deffunc: func() string {
			max := runtime.GOMAXPROCS(-1)
			if max > 1 {
				max = max - 1
			}
			if max < 1 {
				max = 1
			}
			return strconv.Itoa(max)
		},
		doc: "The number of IO bound qtechng commands which are allowed to run in parallel",
	},
	{
		name:      "qtechng-repository-dir",
		mode:      "skip",
		qtechtype: "BP",
		test:      isdir,
		nature:    "dir",
		deffunc: func() string {
			return dirname("/library/repository")
		},
		doc: "Directory on developmentserver and productionservers which contain the Brocade dataset. Should be installed with Ansible.",
	},
	{
		name:      "qtechng-server",
		mode:      "set",
		qtechtype: "WP",
		test:      nil,
		nature:    "dns",
		deffunc: func() string {
			return "dev.anet.be"
		},
		doc: "DNS of the development server. Mainly used in SSH commands.",
	},
	{
		name:      "qtechng-test",
		mode:      "set",
		qtechtype: "BWP",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return "test-entry"
		},
		doc: "Registry entry used for testing purposes.",
	},
	{
		name:      "qtechng-type",
		mode:      "ask",
		qtechtype: "BWP",
		test:      isbwp,
		nature:    "string",
		deffunc: func() string {
			x := qregistry.Registry["qtech-user"]
			if x == "" {
				x = "W"
			}
			return x
		},
		doc: "Working with Brocade, there are 3 types of machines:\n    - `W`: workstations of developers\n    - `B`: the development machine\n    - `P`: production servers",
	},
	{
		name:      "qtechng-user",
		mode:      "ask",
		qtechtype: "BWP",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return qregistry.Registry["qtech-user"]
		},
		doc: "Working with Brocade, there are 3 types of machines:\n    - `W`: workstations of developers\n    - `B`: the development machine\n    - `P`: production servers",
	},
	{
		name:      "qtechng-unique-ext",
		mode:      "ask",
		qtechtype: "B",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return ".m .x"
		},
		doc: "File extensions in Brocade which should be tested on uniqueness.",
	},
	{
		name:      "qtechng-version",
		mode:      "ask",
		qtechtype: "BW",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return "0.00"
		},
		doc: "Default version to use in the development process",
	},
	{
		name:      "qtechng-work-dir",
		mode:      "ask",
		qtechtype: "W",
		test:      isdir,
		nature:    "dir",
		deffunc: func() string {
			return dirname("brocade/work")
		},
		doc: "Local directory containing a version of the Brocade files",
	},
	{
		name:      "qtechng-block-qtechng",
		mode:      "set",
		qtechtype: "B",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return "0"
		},
		doc: "Contains a timestamp. This timestamp prevents remote users to modify the repository",
	},
	{
		name:      "qtechng-block-doc",
		mode:      "set",
		qtechtype: "BP",
		test:      nil,
		nature:    "string",
		deffunc: func() string {
			return "0"
		},
		doc: "Contains a timestamp. This timestamp prevents documentation publishing",
	},
}

func init() {
	registryCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	rootCmd.AddCommand(registryCmd)
}
