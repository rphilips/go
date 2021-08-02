package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
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
	Long: `Gives information about the directory (to be used in shell scripts)
	
The command works with one argument: the name of the directory

Directories are not very important in QtechNG: all information is
about files.

The '--tell' flag specifies which information has to be displayed.
(without this flag, all information is given)
	
'--tell' can have the values:
	
	- dirname: dirname
	- version: version
	- qdir: qdir

The information is infered from the available files in the directory or its
place in the 'qtechng-work-dir'
	
`,
	Example: `qtechng file tell application --cwd=../collections --ext
qtechng dir tell application --cwd=../collections --tell=dirname
qtechng dir tell application --cwd=../collections --tell=qdir
qtechng dir tell application --cwd=../collections --tell=version
qtechng dir tell application --cwd=../collections`,
	Args: cobra.MaximumNArgs(1),
	RunE: dirTell,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	dirTellCmd.Flags().StringVar(&Ftell, "tell", "", "version/qdir/dirname")
	dirCmd.AddCommand(dirTellCmd)
}

func dirTell(cmd *cobra.Command, args []string) error {
	result := make(map[string]string)
	result["version"] = ""
	result["qdir"] = ""
	result["dirname"] = ""

	dirname := Fcwd
	if len(args) != 0 {
		dirname = qutil.AbsPath(args[0], Fcwd)
	}
	result["dirname"] = dirname
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
		result["version"] = qregistry.Registry["qtechng-version"]
	}

	if result["version"] == "" && result["qdir"] != "" && strings.Contains(QtechType, "P") {
		result["version"] = qregistry.Registry["brocade-release"]
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
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}

func dirProps(dirname string) (version string, qdir string) {

	kr := ""
	if strings.Contains(QtechType, "P") {
		kr = qregistry.Registry["brocade-release"]
	}

	if kr == "" && strings.Contains(QtechType, "W") {
		kr = qregistry.Registry["qtechng-version"]
	}
	if kr == "" {
		kr = "0.00"
	}
	dir := new(qclient.Dir)
	dir.Dir = dirname
	m := dir.Repository()
	if len(m) != 0 {
		_, ok := m[kr]
		if ok {
			version = kr
		} else {
			vs := make([]string, len(m))
			i := 0
			for v := range m {
				vs[i] = v
				i++
			}
			sort.Strings(vs)
			version = vs[len(vs)-1]
		}
		qdirs := m[version]
		for d := range qdirs {
			qdir = d
			break
		}
		if len(qdirs) == 1 {
			return version, qdir
		}
	}

	workdir := qregistry.Registry["qtechng-work-dir"]
	if workdir == "" {
		return "", ""
	}
	if !qfs.IsSubDir(workdir, dirname) {
		return version, qdir
	}

	version = kr
	subdir := ""
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
			qdir = ""
			found := false
			parts := strings.SplitN(subdir, "/", -1)
			re := regexp.MustCompile(`^[0-9]+\.[0-9][0-9]$`)
			for _, part := range parts {
				ok := re.MatchString(part)
				if !ok {
					if !found {
						qdir += "/" + part
					}
					continue
				}
				found = true
				version = part
			}
		}
	}

	return version, qdir
}
