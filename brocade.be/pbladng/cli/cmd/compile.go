package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile `pblad`",
	Long:  "Compile `pblad`",

	Args:    cobra.NoArgs,
	Example: `pblad compile`,
	RunE:    compile,
}

func init() {

	rootCmd.AddCommand(compileCmd)
}

func compile(cmd *cobra.Command, args []string) error {

	now := time.Now().Format(time.RFC3339)
	version, err := goversion()
	if err != nil {
		return err
	}
	hostname, err := hostname()
	if err != nil {
		return err
	}
	hostname = strings.ReplaceAll(hostname, " ", "_")
	basedir := pregistry.Registry["source-dir"].(string)
	pkg := pregistry.Registry["package"].(string)
	platforms := pregistry.Registry["platforms"].([]any)
	bexe := pregistry.Registry["exe"].(string)
	bexe = path.Base(bexe)
	if err != nil {
		return err
	}

	for _, platform := range platforms {
		fmt.Println(platform)
		parts := strings.SplitN(platform.(string), "/", -1)
		GOOS := parts[0]
		GOARCH := parts[1]
		os.Setenv("GOOS", GOOS)
		os.Setenv("GOARCH", GOARCH)
		os.Setenv("PBLADNG_BUILDTIME", now)
		os.Setenv("PBLADNG_GOVERSION", version)
		os.Setenv("PBLADNG_BUILDHOST", hostname)

		cwd := filepath.Join(basedir, "brocade.be", "base")
		ptools.Launch([]string{"go", "install", "./..."}, nil, cwd, true, true)
		cwd = filepath.Join(basedir, "brocade.be", pkg, "cli")
		basename := strings.Join([]string{cwd, bexe, GOOS, GOARCH}, "-")
		flags := []string{"-X", "main.buildTime=" + now, "-X", "main.buildHost=" + hostname, "-X", "main.goVersion=" + version}
		params := []string{"go", "build", "-o", basename, "-ldflags", strings.Join(flags, " "), "."}
		out, err := ptools.Launch(params, nil, cwd, true, true)
		sout := string(out)
		if err != nil {
			fmt.Printf("%s %s at %s:\n%s\n\nargs:%v", basename, err.Error(), cwd, sout, params)
			continue

		}
		if GOOS == runtime.GOOS {
			exe := pregistry.Registry["exe"].(string)
			pexe, _ := exec.LookPath(exe)
			err = bfs.RefreshEXE(pexe, basename)
			if err != nil {
				return err
			}
			if GOOS == "linux" {
				err = bfs.SetPathmode(pexe, "scriptfile")
				if err != nil {
					return err
				}
				perm := bfs.CalcPerm("rwxrwxr-x")
				err = os.Chmod(pexe, perm|os.ModeSetuid|os.ModeSetgid)
				if err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func goversion() (string, error) {
	args := []string{"go", "version"}
	out, err := ptools.Launch(args, nil, "", true, true)
	sout := string(out)
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), sout)
		return "", err
	}
	return strings.SplitN(sout, " ", -1)[2], nil
}

func hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return hostname, err
}
