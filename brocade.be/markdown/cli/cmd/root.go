package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	pfs "brocade.be/base/fs"
)

// BuildTime defined by compilation
var BuildTime = ""

// GoVersion defined by compilation
var GoVersion = ""

// BuildHost defined by compilation
var BuildHost = ""

// Fcwd Current working directory
var Fcwd string // current working directory

// Fenv environment variables
var Fenv []string

// Fsilent if true: no output
var Fsilent bool

// Fstdout if not-empty: output is redirected to this file
var Fstdout string

// Fmsg output of commands
var Fmsg string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "markdown",
	Short:         "Pblad executive",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `markdown maintains the Pblad software:

    - Development
    - Installation
	- Deployment
	- Management`,
	PersistentPreRunE: preRun,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&Fstdout, "stdout", "", "Filename containing the result")
	rootCmd.PersistentFlags().StringVar(&Fcwd, "cwd", "", "Working directory")
	rootCmd.PersistentFlags().BoolVar(&Fsilent, "quiet", false, "Silent the output")
	rootCmd.PersistentFlags().StringArrayVar(&Fenv, "env", []string{}, "Environment variable KEY=VALUE")

}

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

	// Cwd
	Fcwd, err = checkCwd(Fcwd)
	if err != nil {
		return
	}

	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(buildTime string, goVersion string, buildHost string, args []string) {
	BuildTime = buildTime
	BuildHost = buildHost
	GoVersion = goVersion
	err := rootCmd.Execute()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if Fmsg != "" && !Fsilent {
		if Fstdout == "" {
			fmt.Println(Fmsg)
			return
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		fmt.Fprintln(w, Fmsg)
		err = w.Flush()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func checkCwd(cwd string) (dir string, err error) {
	defer func() { dir, _ = pfs.AbsPath(dir) }()
	dir = cwd
	if cwd == "" {
		dir, err = os.Getwd()
		if err != nil {
			return
		}
	}
	if !pfs.Exists(dir) || !pfs.IsDir(dir) {
		err = fmt.Errorf("`%s` does not exist or is not a directory", dir)
		return
	}
	return
}