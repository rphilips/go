package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	ptools "brocade.be/pbladng/lib/tools"

	"github.com/spf13/cobra"
)

var Fext = ""
var Fmatch = ""
var Fsort = ""
var Fname = ""
var Freport = ""

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "File manipulation",
	Long:  `Manipulate file names`,

	Example: `gopblad rename rename`,
	RunE:    rename,
}

func init() {
	renameCmd.PersistentFlags().StringVar(&Fext, "ext", "jpg,jpeg", "work on this list of comma separated extension")
	renameCmd.PersistentFlags().StringVar(&Fsort, "sort", "name", "sort on name, time or partnummer: separate by comma")
	renameCmd.PersistentFlags().StringVar(&Fmatch, "match", "", "regular expression within ^...$")
	renameCmd.PersistentFlags().StringVar(&Fname, "name", "", "new name: a replacement template")
	renameCmd.PersistentFlags().StringVar(&Freport, "report", "", "identifier of report section")
	rootCmd.AddCommand(renameCmd)
}

func rename(cmd *cobra.Command, args []string) error {
	if Fname == "" {
		return fmt.Errorf("should specify a `--name=...` flag")
	}
	Fmatch = strings.TrimPrefix(Fmatch, "^")
	Fmatch = strings.TrimSuffix(Fmatch, "$")
	if Fmatch == "" {
		Fmatch = "(.*)"
	}
	Fmatch = "^" + Fmatch + "$"
	pattern, err := regexp.Compile(Fmatch)

	if err != nil {
		return fmt.Errorf("invalid regexp for `--match=...` flag")
	}
	Freport := strings.TrimSpace(Freport)

	reportnr := -1
	if Freport != "" {
		var e error
		reportnr, e = strconv.Atoi(Freport)
		if e != nil {
			return fmt.Errorf("invalid reportnumber for `--report=...` flag")
		}
	}

	if Fdir == "" {
		Fdir = "."
	}
	files, _, err := bfs.FilesDirs(Fdir)

	if err != nil {
		return err
	}
	rev := make(map[string]bool)
	for _, f := range files {
		rev[f.Name()] = true
	}

	work := make([]os.FileInfo, 0, len(files))
	if len(args) != 0 {
		for _, f := range files {
			ok := false
			for _, arg := range args {
				ok = arg == f.Name()
				if ok {
					work = append(work, f)
					break
				}
			}
		}
	}
	if len(work) == 0 {
		return nil
	}
	files = work[:]

	work = make([]os.FileInfo, 0, len(files))
	if Fext != "-" {
		for _, f := range files {
			ext := filepath.Ext(f.Name())
			ext = strings.ToLower(ext)
			ok := false
			for _, piece := range strings.SplitN(Fext, ",", -1) {
				piece := "." + strings.TrimLeft(piece, ".")
				piece = strings.ToLower(strings.TrimSpace(piece))
				ok = ext == piece
				if ok {
					work = append(work, f)
					break
				}
			}
		}
	}

	if len(work) == 0 {
		return nil
	}

	files = work[:]
	work = make([]os.FileInfo, 0, len(files))
	m := make(map[string][][]string)

	for _, f := range files {
		base := filepath.Base(f.Name())
		subs := pattern.FindAllStringSubmatch(base, -1)
		if len(subs) == 0 {
			continue
		}
		m[f.Name()] = subs
		work = append(work, f)
	}

	if len(m) == 0 {
		return nil
	}
	if strings.Contains(Fname, "$#") {

		fslice := make([]func(int, int) int, 0)
		for _, part := range strings.SplitN(Fsort, ",", -1) {
			part = strings.ToLower(strings.TrimSpace(part))
			if part == "" {
				continue
			}
			switch {
			case part == "name":
				f := func(i, j int) int {
					return strings.Compare(strings.ToLower(work[i].Name()), strings.ToLower(work[j].Name()))
				}
				fslice = append(fslice, f)
			case part == "time":
				f := func(i, j int) int {
					if work[i].ModTime().Before(work[j].ModTime()) {
						return -1
					} else {
						return 1
					}
				}
				fslice = append(fslice, f)
			default:
				k, e := strconv.Atoi(part)
				if e != nil || k == 0 {
					return fmt.Errorf("invalid sort indicator `%s`", part)
				}
				f := func(i, j int) int {
					subsi := m[work[i].Name()]
					vi := ""
					if len(subsi) > k {
						vi = subsi[0][k]
					}
					subsj := m[work[j].Name()]
					vj := ""
					if len(subsj) > k {
						vj = subsj[0][k]
					}
					return strings.Compare(vi, vj)
				}
				fslice = append(fslice, f)
			}
		}
		less := func(i, j int) bool {
			for _, f := range fslice {
				r := f(i, j)
				if r == 0 {
					continue
				}
				if r == -1 {
					return true
				}
				return false
			}
			return false
		}
		sort.Slice(work, less)
	}
	renames := make(map[string]string)
	reports := make(map[string]string)

	frame := fmt.Sprintf("%%0%dd", len(strconv.Itoa(len(work))))
	for k, f := range work {
		template := Fname
		if strings.Contains(template, "$#") {
			template = strings.ReplaceAll(template, "$#", fmt.Sprintf(frame, k+1))
		}
		subs := m[f.Name()]
		n := len(subs[0]) - 1
		for {
			if n < 1 {
				break
			}
			sub := subs[0][n]
			template = strings.ReplaceAll(template, "$"+strconv.Itoa(n), sub)
			n--
		}
		result := []byte{}
		sresult := string(pattern.ExpandString(result, template, f.Name(), pattern.FindAllStringSubmatchIndex(f.Name(), -1)[0]))
		if rev[sresult] {
			return fmt.Errorf("`%s` renames to `%s` but this exists", f.Name(), sresult)
		}
		rev[sresult] = true
		renames[f.Name()] = sresult
		report := ""
		rnr := -1
		if Freport == "" {
			rnr = len(subs[0]) - 1
		} else {
			rnr = reportnr
		}
		if len(subs[0]) > rnr {
			report = subs[0][rnr]
		}
		reports[f.Name()] = strings.TrimSpace(report)
	}

	maxo := 0
	maxn := 0
	for old, new := range renames {
		if len(old) > maxo {
			maxo = len(old)
		}
		if len(new) > maxn {
			maxn = len(new)
		}
	}
	frame = "%" + strconv.Itoa(maxo) + "s -> %" + strconv.Itoa(maxn) + "s %s\n"
	for _, oldf := range work {
		old := oldf.Name()
		new := renames[old]

		fmt.Printf(frame, old, new, reports[old])

		err := os.Rename(old, renames[old])
		if err != nil {
			return fmt.Errorf("`%s` renames to `%s`: error %s", old, renames[old], err)
		}
		fmt.Println(old, "->", renames[old])
	}
	ask := "\nRename ? (y/n): "
	rn := ptools.YesNo(ask)

	if !rn {
		return nil
	}

	for _, oldf := range work {
		old := oldf.Name()
		new := renames[old]

		err := os.Rename(old, new)
		if err != nil {
			return fmt.Errorf("`%s` renames to `%s`: error %s", old, renames[old], err)
		}
		fmt.Printf(frame, old, new, reports[old])
	}

	return nil
}
