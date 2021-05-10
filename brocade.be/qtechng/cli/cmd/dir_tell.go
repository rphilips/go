package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

//Ftell tells what kind of informatiom has to be returned
var dirTellCmd = &cobra.Command{
	Use:   "tell",
	Short: "Gives information about the directory",
	Long:  `Gives information about the directory`,
	Example: `  qtechng file tell /home/rphilips/catalografie --cwd=../catalografie --ext
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=dirname
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=project
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=qdir
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=version
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=abspath
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie --tell=relpath
	  qtechng dir tell /home/rphilips/catalografie --cwd=../catalografie
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
	result := make(map[string]string)
	result["version"] = ""
	result["qdir"] = ""

	dirname := Fcwd
	if len(args) != 0 {
		dirname = qutil.AbsPath(args[0], Fcwd)
	}
	dir := new(qclient.Dir)
	dir.Dir = dirname
	m := dir.Repository()
	if m != nil {
		if len(m) == 1 {
			for r := range m {
				result["version"] = r
				break
			}
		}
		qdirs := make(map[string]bool)
		for r := range m {
			for s := range m[r] {
				if s != "" {
					qdirs[s] = true
				}
			}
		}
		if len(qdirs) == 1 {
			for s := range qdirs {
				result["qdir"] = s
				break
			}
		}
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
			subdir = strings.TrimPrefix(subdir, "./")
			if subdir == "." {
				subdir = ""
			}
			subdir = "/" + subdir
			first := ""
			last := ""
			parts := strings.SplitN(subdir, "/", -1)
			re := regexp.MustCompile(`^[0-9]+\.[0-9][0-9]$`)
			for _, part := range parts {
				ok := re.MatchString(part)
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
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml)
	return nil
}
