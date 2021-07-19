package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qlfile "brocade.be/qtechng/lib/file/lfile"
	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var textImportCmd = &cobra.Command{
	Use:   "import",
	Short: "Import lgcodes",
	Long: `Command which import *lgcodes*.
There is only one argument: the name of the import file.
This file should be created by a matching *export* command.

The entries are checked and, if valid, imported in the corresponding L-files.
An extra column is added with a status message.

Different error messages:

    - OK: no error found
	- MISSING: the lgcode is unknown
	- INCOMPLETE: incomplete record
	- READFILE: cannot read lgcode
	- LFILE: missing L-file
	- UNCHANGED: data is not changed
	- OUTDATED: the translation is not valid (the original text is changed)
`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng text export N E F`,

	RunE:   textImport,
	PreRun: preImport,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

func init() {
	textImportCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	textImportCmd.Flags().StringVar(&Flist, "list", "", "List with imported file and additional information")
	textCmd.AddCommand(textImportCmd)
}

func preImport(cmd *cobra.Command, args []string) {
	if Ftransported {
		return
	}
	csvfile := args[0]
	blob, e := os.ReadFile(csvfile)
	if e != nil {
		err := qerror.QError{
			Ref:  []string{"import.read.file"},
			File: csvfile,
			Msg:  []string{"`" + csvfile + "` read with error: " + e.Error()},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		Fpayload = nil
		return
	}
	transports := make([]qclient.Transport, 0)
	transports = append(transports, qclient.Transport{
		LocFile: qclient.LocalFile{},
		Body:    blob,
	})

	Fpayload = &qclient.Payload{
		ID:         "Once",
		UID:        FUID,
		CMD:        "qtechng",
		Origin:     QtechType,
		Args:       os.Args[1:],
		Transports: transports,
	}
	ok := false
	for _, arg := range Fpayload.Args {
		if strings.HasPrefix(arg, "--version=") {
			ok = true
			break
		}
	}
	if !ok {
		Fpayload.Args = append(Fpayload.Args, "--version="+Fversion)
	}

	if !strings.ContainsRune(QtechType, 'B') {
		whowhere := qregistry.Registry["qtechng-server"]
		if !strings.Contains(whowhere, "@") {
			whowhere = FUID + "@" + whowhere
		}
		catchOut, catchErr, err := qssh.SSHcmd(Fpayload, whowhere)
		if err != nil {
			log.Fatal("cmd/text_import/preImported/1:\n", err)
		}
		if catchErr.Len() != 0 {
			log.Fatal("cmd/text_import/preImported/2:\n", catchErr)
		}
		fmt.Print(catchFile(catchOut.String(), Flist))
	}

}

func textImport(cmd *cobra.Command, args []string) error {
	if !Ftransported && Fpayload == nil {
		return nil
	}
	csvfile := ""
	version := Fversion
	if strings.ContainsRune(QtechType, 'B') {
		filename := Fpayload.Args[2]
		filename = filepath.ToSlash(filename)
		csvfile = filename
		k := strings.LastIndex(filename, "/")
		if k != -1 {
			csvfile = filename[k+1:]
		}
		csvfile = filepath.Join(qregistry.Registry["scratch-dir"], csvfile)
		err := qfs.Store(csvfile, Fpayload.Transports[0].Body, "qtech")
		if err != nil {
			return err
		}
		for _, arg := range Fpayload.Args {
			if strings.HasPrefix(arg, "--version=") {
				x := strings.TrimPrefix(arg, "--version=")
				if x != "" {
					version = x
					break
				}
			}
		}
		csvfile, numbers, backupdir, err := handleCSV(version, csvfile)

		result := make(map[string]string)

		if err == nil {
			result["#filename"] = csvfile
			result["#backupdir"] = backupdir

			for x, nr := range numbers {
				result[x] = strconv.Itoa(nr)
			}
		} else {
			result = nil
		}

		Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}

	return nil

}

func handleCSV(version string, csvfile string) (filename string, numbers map[string]int, backupdir string, err error) {

	f, err := os.Open(csvfile)
	if err != nil {
		return
	}
	first, err := bufio.NewReader(f).ReadString('\n')
	f.Close()
	if err != nil && err != io.EOF {
		return
	}
	if strings.TrimSpace(first) == "" {
		err = errors.New("missing first line")
		return
	}
	k := strings.Count(first, ",")
	l := strings.Count(first, ";")
	delim := ';'
	if l < k {
		delim = ','
	}
	crlf := strings.ContainsRune(first, '\r')
	first = strings.TrimRight(first, "\r\n")
	pieces := strings.SplitN(first, string(delim), -1)
	if len(pieces) < 5 {
		err = errors.New("first line should have at least 5 fields")
		return
	}
	if pieces[0] != "source" {
		err = errors.New("first line should have field `source`")
		return
	}
	if pieces[1] != "code" {
		err = errors.New("first line should have field `code`")
		return
	}
	len := len(pieces)
	lgs := make([]string, 0)
	baselg := ""
	m := make(map[string]bool)
	for i := 2; i < len-1; i++ {
		code := pieces[i]
		if m[code] {
			err = fmt.Errorf("`%s`: twice in the header", code)
			return
		}
		m[code] = true
		ok := false
		for _, lg := range []string{"dut", "fre", "eng", "ger"} {
			if lg == code {
				lgs = append(lgs, lg)
				ok = true
				if baselg == "" {
					baselg = lg
				}
				break
			}
		}
		if !ok {
			err = fmt.Errorf("`%s`: not a language", code)
			return
		}
	}
	if baselg == "" {
		err = errors.New("no languages in header")
		return
	}
	f, err = os.Open(csvfile)
	if err != nil {
		return
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = delim
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	records, err := r.ReadAll()
	if err != nil {
		return
	}

	// handle records

	filename, numbers, backupdir, errs := rewriteCVS(version, records, baselg, lgs, crlf, delim)
	if errs == nil {
		return csvfile, numbers, backupdir, nil
	}
	return filename, numbers, backupdir, qerror.ErrorSlice(errs)
}

func rewriteCVS(v string, records [][]string, baselg string, lgs []string, crlf bool, delim rune) (filename string, numbers map[string]int, backupdir string, errs []error) {
	if len(records) == 0 {
		return "", nil, "", []error{errors.New("no translations found")}
	}
	version, err := qserver.Release{}.New(v, true)
	if err != nil {
		return "", nil, "", []error{err}
	}

	lgmap := make(map[string]int)
	index := 2 + len(lgs)
	indexlg := -1
	for i, lgi := range records[0] {
		for _, lg := range lgs {
			if lg == lgi {
				lgmap[lgi] = i
			}
		}
	}
	indexlg = lgmap[baselg]

	fn := func(n int) (interface{}, error) {
		record := records[n]
		if n == 0 {
			record[index] = "status"
			return record, nil
		}
		for len(record) <= index {
			record = append(record, "")
		}
		code := record[1]
		if code == "" {
			record[index] = "INCOMPLETE"
			return record, nil
		}
		fs, place := version.ObjectPlace("l4_" + code)

		ok, e := fs.Exists(place)
		if !ok || e != nil {
			record[index] = "MISSING"
			return record, nil
		}
		data, e := fs.ReadFile(place)
		if e != nil {
			record[index] = "READFILE"
			return record, nil
		}
		m := make(map[string]string)
		e = json.Unmarshal(data, &m)
		if e != nil {
			record[index] = "READFILE"
			return record, nil
		}
		source := m["source"]
		if source != "" {
			record[0] = source
		}
		source = record[0]

		_, e = qsource.Source{}.New(v, source, true)
		if e != nil {
			record[index] = "LFILE"
			return record, nil
		}

		base := m[baselg]
		if record[indexlg] != base {
			record[index] = "OUTDATED:" + base
			return record, nil
		}

		ok = true
		for _, lg := range lgs {
			i := lgmap[lg]
			if record[i] != m[lg] {
				ok = false
				break
			}
		}
		if ok {
			record[index] = "UNCHANGED"
			return record, nil
		}
		return record, nil
	}

	result, _ := qparallel.NMap(len(records), -1, fn)
	for i, r := range result {
		records[i] = r.([]string)
	}

	// find l-files
	mlfiles := make(map[string][]int)
	lfiles := make([]string, 0)
	for i, record := range records {
		if i == 0 {
			continue
		}
		if record[index] != "" {
			continue
		}
		source := record[0]
		inxes := mlfiles[source]
		if len(inxes) == 0 {
			inxes = make([]int, 0)
			lfiles = append(lfiles, source)
		}
		mlfiles[source] = append(inxes, i)
	}

	// handle lfiles
	stamp := time.Now().Format(time.RFC3339)
	meta := qmeta.Meta{
		Fu: "translation",
		Ft: stamp,
	}
	stamp = strings.ReplaceAll(stamp, ":", "")
	stamp = strings.ReplaceAll(stamp, "-", "")

	// backup lfiles
	if len(lfiles) != 0 {
		backupdir, _ = qfs.TempDir("", "lfiles-")
		args := []string{
			"source",
			"co",
			"--version=" + v,
			"--tree",
			"--copyonly",
		}
		args = append(args, lfiles...)
		qutil.QtechNG(args, "", false, backupdir)
	}

	fl := func(n int) (interface{}, error) {
		qpath := lfiles[n]
		source, _ := qsource.Source{}.New(v, qpath, false)
		blob, err := source.Fetch()
		if err != nil {
			return nil, err
		}
		lfile := new(qlfile.LFile)
		lfile.SetEditFile(qpath)
		lfile.SetRelease(v)
		err = qobject.Loads(lfile, blob, true)
		if err != nil {
			return nil, err
		}
		objs := lfile.Objects()
		lfile.SetObjects(objs)
		indexes := make(map[string]int)
		for _, i := range mlfiles[qpath] {
			code := records[i][1]
			indexes[code] = i
		}
		for i, lgcode := range lfile.Lgcodes {
			code := lgcode.ID
			inx, ok := indexes[code]
			if !ok {
				continue
			}
			record := records[inx]
			for _, lg := range lgs {
				j := lgmap[lg]
				newtr := record[j]
				switch lg {
				case "dut":
					lgcode.N = newtr
				case "eng":
					lgcode.E = newtr
				case "fre":
					lgcode.F = newtr
				case "ger":
					lgcode.D = newtr
				case "unv":
					lgcode.U = newtr
				}
			}
			lfile.Lgcodes[i] = lgcode
		}
		output := lfile.Format()
		_, _, _, err = source.Store(meta, output.Bytes(), false)
		return nil, err
	}

	_, errlist := qparallel.NMap(len(lfiles), -1, fl)
	errs = make([]error, 0)
	for _, e := range errlist {
		if e != nil {
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		errs = nil
	}

	numbers = make(map[string]int)
	for i, record := range records {
		if i == 0 {
			continue
		}
		status := record[index]
		if status == "" {
			status = "IMPORTED"
		}
		if strings.ContainsRune(status, ':') {
			status = strings.SplitN(status, ":", 2)[0]
		}
		numbers[status] += 1
	}
	csvfile := qutil.AbsPath("brocade-"+v+"-"+stamp+".csv", qregistry.Registry["scratch-dir"])
	f, err := os.Create(csvfile)
	if err != nil {
		return "", numbers, backupdir, []error{err}
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = ';'
	w.UseCRLF = true
	w.WriteAll(records) // calls Flush internally
	err = w.Error()
	if err != nil {
		return "", numbers, backupdir, []error{err}
	}
	return csvfile, numbers, backupdir, nil
}
