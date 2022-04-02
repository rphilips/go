package cmd

import (
	"encoding/hex"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// BuildTime defined by compilation
var BuildTime = ""

// GoVersion defined by compilation
var GoVersion = ""

// BuildHost defined by compilation
var BuildHost = ""

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(buildTime string, goVersion string, buildHost string, args []string) {
	BuildTime = buildTime
	BuildHost = buildHost
	GoVersion = goVersion
	rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:               "iiiftool",
	Short:             "iiiftool - CLI application to handle IIIF",
	SilenceUsage:      true,
	SilenceErrors:     true,
	Long:              `iiiftool is a CLI application to handle IIIF`,
	PersistentPreRunE: preRun,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&Funhex, "unhex", false, "Unhexify the arguments starting with `.`")
	rootCmd.PersistentFlags().StringVar(&Fcwd, "cwd", "", "Working directory")
}

//Fenv environment variables
var Fenv []string

//Fsilent environment variables
var Fsilent bool

// Funhex decides if the args are to be unhexed (if starting with `.`)
var Funhex bool

// Fcwd is the working directory for the command
var Fcwd string

// Fquality is the quality parameter for image conversion
var Fquality int

// Ftile is the tiling parameter for image conversion
var Ftile int

// Fverbose is the parameter for verbose processing
var Fverbose bool

func preRun(cmd *cobra.Command, args []string) (err error) {

	if len(Fenv) != 0 {
		for _, env := range Fenv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 0 {
				continue
			}
			env := strings.TrimSpace(parts[0])
			if env == "" {
				continue
			}
			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}
			os.Setenv(env, value)
		}

	}

	for count, arg := range args {
		if strings.HasPrefix(arg, ".") {
			arg = arg[1:]
			decoded, e := hex.DecodeString(arg)
			if e == nil {
				args[count] = string(decoded)
			}
		}
	}

	if err != nil {
		return
	}

	return
}
