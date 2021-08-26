package cmd

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var textExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export lgcodes",
	Long: `This command exports *lgcodes* (terms which can be translated).
All lgcodes are checked and, if appropriate, are written to a CSV file.

The arguments are the languages which should be exported.
If no arguments are given, the languages "eng", "fre", "dut" are included.

The flag '--emptyonly' selects only those lgcodes from which a translation is
missing.

The end result is a file which is copied to the 'download' subdirectory of
'qtechng-work-dir'.
`,
	Args:    cobra.MaximumNArgs(0),
	Example: `qtechng text export`,

	RunE: textExport,
	PreRun: func(cmd *cobra.Command, args []string) {
		preSSH(cmd, func(result string) string { return catchFile(result, Flist) })
	},
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
		"fill-version":      "yes",
	},
}

var Femptyonly bool

func init() {
	textExportCmd.Flags().BoolVar(&Femptyonly, "emptyonly", false, "Export only if a translation is empty")
	textExportCmd.Flags().StringVar(&Flist, "list", "", "List with exported file")
	textExportCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	textCmd.AddCommand(textExportCmd)
}

func textExport(cmd *cobra.Command, args []string) error {
	lgs := make([]string, 0)
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		arg = strings.Replace(arg, "_", "-", -1)
		arg = strings.Replace(arg, ":", "-", -1)
		arg = strings.Replace(arg, ".", "-", -1)
		arg = strings.Replace(arg, " ", "-", -1)
		if strings.ContainsRune(arg, '-') {
			arg = strings.SplitN(arg, "-", -1)[0]
		}
		arg = strings.ToLower(arg)
		switch arg {
		case "n", "nl", "ned":
			lgs = append(lgs, "dut")
		case "f", "fre", "fr":
			lgs = append(lgs, "fre")
		case "e", "eng", "en":
			lgs = append(lgs, "eng")
		case "d", "ger", "ge":
			lgs = append(lgs, "ger")
		default:
			lgs = append(lgs, arg)
		}
	}
	if len(lgs) == 0 {
		lgs = []string{"dut", "fre", "eng"}
	}

	url, numbers, err := lgsURL(Fversion, lgs, Femptyonly)

	result := make(map[string]string)

	if err == nil {

		result["#filename"] = url

		for lg, nr := range numbers {
			result[lg] = strconv.Itoa(nr)
		}
	} else {
		result = nil
	}

	Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")

	return nil

}

func lgsURL(v string, lgs []string, emptyonly bool) (url string, numbers map[string]int, err error) {
	sort.Strings(lgs)

	version, err := qserver.Release{}.New(v, true)
	if err != nil {
		return
	}
	fs := version.FS("/")
	dir, _ := fs.RealPath("/object/l4")
	allfiles, _ := qfs.Find(dir, []string{"obj.json"}, true, true, false)

	fn := func(n int) (interface{}, error) {
		file := allfiles[n]
		data, e := qfs.Fetch(file)
		if e != nil {
			return nil, e
		}
		m := make(map[string]string)
		e = json.Unmarshal(data, &m)
		if e != nil {
			return nil, e
		}
		skip := true
		for _, lg := range lgs {
			if m[lg] != "" {
				skip = false
				break
			}
		}
		if skip {
			return nil, nil
		}
		if emptyonly {
			ok := false
			for _, lg := range lgs {
				if m[lg] == "" {
					ok = true
					break
				}
			}
			if !ok {
				return nil, nil
			}
		}
		return m, nil
	}
	result, _ := qparallel.NMap(len(allfiles), -1, fn)

	records := make([][]string, 0)
	h := time.Now()
	t := h.Format(time.RFC3339)[:19]
	t = strings.ReplaceAll(t, ":", "")
	t = strings.ReplaceAll(t, "-", "")
	stamp := t
	header := []string{"source", "code"}

	header = append(header, lgs...)
	header = append(header, stamp)
	records = append(records, header)

	numbers = make(map[string]int)
	count := 0
	empty := 0
	for _, r := range result {
		if r == nil {
			continue
		}
		count++
		m := r.(map[string]string)
		row := []string{m["source"], m["id"]}
		for _, lg := range lgs {
			x := m[lg]
			row = append(row, x)
			if x == "" {
				numbers[lg+"?"] += 1
				empty += 1
			} else {
				numbers[lg+"!"] += 1
			}
		}
		records = append(records, row)
	}
	numbers["#total"] = count
	numbers["#untranslated"] = empty

	csvfile := qutil.AbsPath("brocade-"+v+"-"+t+".csv", qregistry.Registry["scratch-dir"])
	f, err := os.Create(csvfile)
	if err != nil {
		return
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = ';'
	w.UseCRLF = true
	w.WriteAll(records) // calls Flush internally
	err = w.Error()
	if err != nil {
		return
	}
	url = csvfile
	return

}

func catchFile(result string, editlist string) string {

	if result == "" {
		return result
	}
	filename, e := qutil.JSONpath([]byte(result), "$..#filename")
	if e != nil {
		return result
	}
	filename = strings.TrimSpace(filename)
	if filename == "" {
		return result
	}
	if strings.HasPrefix(filename, "{") {
		return result
	}
	if !strings.HasPrefix(filename, `"`) {
		return result
	}

	replace := filename
	s := ""
	json.Unmarshal([]byte(filename), &s)
	filename = s
	if filename == "" {
		return result
	}
	s = filepath.ToSlash(filename)
	k := strings.LastIndex(s, "/")
	base := s
	if k != -1 {
		base = s[k+1:]
	}

	dir := qregistry.Registry["scratch-dir"]
	if strings.ContainsRune(qregistry.Registry["qtechng-type"], 'W') && qregistry.Registry["qtechng-support-dir"] != "" && qregistry.Registry["qtechng-work-dir"] != "" {
		dir = filepath.Join(qregistry.Registry["qtechng-work-dir"], "download")
		qfs.MkdirAll(dir, "qtech")
		qutil.EditList(editlist, false, []string{"/download/" + base})
	}
	localfile := filepath.Join(dir, base)
	//url := qutil.FileURL(localfile, -1)
	bs, _ := json.Marshal(localfile)
	result = strings.Replace(result, replace, string(bs), -1)
	data, _ := ReadSSHAll(filename)

	qfs.Store(localfile, data, "qtech")

	return result
}
