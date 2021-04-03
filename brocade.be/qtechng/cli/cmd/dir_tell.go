package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

//Ftell tells what kind of informatiom has to be returned
var dirTellCmd = &cobra.Command{
	Use:   "tell",
	Short: "Gives information about the directory",
	Long:  `Gives information about the directory`,
	Example: `  qtechng file tell bcawedit.m --cwd=../catalografie --ext
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=dirname
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=project
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=qdir
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=version
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=abspath
	  qtechng dir tell bcawedit.m --cwd=../catalografie --tell=relpath
	  qtechng dir tell bcawedit.m --cwd=../catalografie
	`,
	Args: cobra.MaximumNArgs(1),
	RunE: dirTell,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	dirTellCmd.Flags().StringVar(&Ftell, "tell", "", "version/qdir")
	dirCmd.AddCommand(dirTellCmd)
}

func dirTell(cmd *cobra.Command, args []string) error {
	dirname := Fcwd
	if len(args) != 0 {
		dirname := args[0]
		if !filepath.IsAbs(dirname) {
			dirname = path.Join(Fcwd, dirname)
		}
	}
	dir := new(qclient.Dir)
	dir.Dir = dirname
	dir.Load()
	result := make(map[string]string)
	releases := make(map[string]bool)
	qdirs := make(map[string]bool)

	release := ""
	qdir := ""
	for _, locfil := range dir.Files {
		r := locfil.Release
		if r == "" {
			continue
		}
		release = r
		qpath := locfil.QPath
		if qpath == "" {
			continue
		}
		releases[r] = true
		q, _ := qutil.QPartition(qpath)
		qdirs[q] = true
		qdir = q
	}

	result["version"] = ""
	result["qdir"] = ""
	if len(releases) == 1 {
		result["version"] = release
	}
	if len(qdirs) == 1 {
		result["qdir"] = qdir
	}
	workdir := qregistry.Registry["qtechng-work-dir"]
	subdir := ""
	subrel := ""
	if workdir != "" {
		workdir, _ = qfs.AbsPath(workdir)
		dirname, _ = qfs.AbsPath(dirname)
		relpath, _ := filepath.Rel(workdir, dirname)
		if !strings.HasPrefix(relpath, "..") {
			subdir = filepath.ToSlash(relpath)
			if strings.HasPrefix(subdir, "./") {
				subdir = subdir[2:]
			}
			if subdir == "." {
				subdir = ""
			}
			subdir = "/" + subdir
			first := ""
			last := ""
			parts := strings.SplitN(subdir, "/", -1)
			for _, part := range parts {
				ok, _ := regexp.MatchString(`^[0-9][0-9]?\.[0-9][0-9]$`, part)
				if !ok {
					continue
				}
				if first == "" {
					first = part
				}
				last = part
			}
			subrel = last
			if first != "" {
				subdir = subdir + "/"
				subdir = strings.SplitN(subdir, "/"+first+"/", -1)[0]
			}
		}
	}
	if result["qdir"] == "" {
		result["qdir"] = subdir
	}
	if result["version"] == "" {
		result["version"] = subrel
	}

	if result["version"] == "" && result["qdir"] != "" && !strings.Contains(QtechType, "P") {
		result["version"] = "0.00"
	}

	tell, ok := result[Ftell]

	if ok {
		if Fstdout == "" || Ftransported {
			fmt.Print(tell)
			return nil
		}
		f, err := os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		fmt.Fprint(w, tell)
		err = w.Flush()
		return err
	}
	Fmsg = qerror.ShowResult(result, Fjq, nil, Fyaml)
	return nil
}
