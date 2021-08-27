package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var objectMoveCmd = &cobra.Command{
	Use:   "move [objects]",
	Short: "Move objects to a source file",
	Long: `This command moves objects to a source file
The objects can be specified:

    - as arguments
	- as '--objpattern-...' flags.

Do not forget the appropriate prefix!

The *receiving* objectfile is specified by the '--receiver=...' flag:
it contains its 'qpath'.`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng object move m4_getPkObject m4_getPkLib --receiver=/cats/application/description.d
qtechng object move --objpattern='m4_getCat*' --receiver=/cats/application/description.d
	`,
	RunE: objectMove,
	PreRun: func(cmd *cobra.Command, args []string) {
		preSSH(cmd, nil)
	},
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"with-qtechtype":    "BW",
		"fill-version":      "yes",
	},
}

var Freceiver string

func init() {
	objectCmd.AddCommand(objectMoveCmd)
	objectMoveCmd.PersistentFlags().StringArrayVar(&Fobjpattern, "objpattern", []string{}, "Posix glob pattern on object names")
	objectMoveCmd.PersistentFlags().StringVar(&Freceiver, "receiver", "", "receiving object file")
}

func objectMove(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		if len(Fobjpattern) == 0 {
			Fobjpattern = args
		} else {
			for _, arg := range args {
				ok := len(Fobjpattern) == 0
				for _, p := range Fobjpattern {
					if p == arg {
						ok = true
						break
					}
					if qutil.EMatch(p, arg) {
						ok = true
						break
					}
				}
				if ok {
					Fobjpattern = append(Fobjpattern, arg)
				}
			}
		}
	}
	if len(Fobjpattern) == 0 {
		err := errors.New("no objects specified")
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	ext := path.Ext(Freceiver)
	if ext != ".d" && ext != ".l" && ext != ".i" {
		err := fmt.Errorf("`%s` has the wrong extension", Freceiver)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	ty := ""
	switch ext {
	case ".d":
		ty = "m4"
	case ".l":
		ty = "l4"
	case ".i":
		ty = "i4"
	}
	_, err := qserver.Release{}.New(Fversion, true)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	source, err := qsource.Source{}.New(Fversion, Freceiver, true)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	natures := source.Natures()
	if !natures["objectfile"] {
		err := fmt.Errorf("`%s` is NOT an object file", Freceiver)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	objs := qobject.FindObjects(Fversion, Fobjpattern)

	if len(objs) == 0 {
		err := errors.New("no matching objects found")
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	wobjs := make([]string, 0)
	sources := make(map[string][]string)
	oldsources := make(map[string]string)

	for _, obj := range objs {
		if !strings.HasPrefix(obj, ty) {
			err := fmt.Errorf("`%s` is of the wrong type", obj)
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		oldsource := qobject.GetEditFile(Fversion, obj)
		oldsources[obj] = oldsource
		if oldsource == Freceiver {
			continue
		}
		x := sources[oldsource]
		if len(x) == 0 {
			x = make([]string, 0)
		}
		x = append(x, obj)
		sources[oldsource] = x
		wobjs = append(wobjs, obj)
	}

	if len(wobjs) == 0 {
		Fmsg = qreport.Report(objs, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	tmpdir, err := qfs.TempDir("", "objmove.")
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	retrsources := []string{Freceiver}
	for _, q := range oldsources {
		retrsources = append(retrsources, q)
	}
	argums := []string{"source", "co", "--version=" + Fversion, "--tree"}
	argums = append(argums, retrsources...)
	_, serr, err := qutil.QtechNG(argums, []string{"$..ERROR"}, false, tmpdir)

	if serr != "" {
		err = fmt.Errorf("checkout of relevant sources gives error: `%s`", serr)
	}
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	// extract the data matching the objects
	objcontents := make(map[string]string)
	fn1 := func(n int) (interface{}, error) {
		source := retrsources[n]
		if source == Freceiver {
			return nil, nil
		}
		objs := sources[source]
		content := extractDefinition(tmpdir, source, objs)
		return content, nil
	}
	rlist, _ := qparallel.NMap(len(retrsources), -1, fn1)

	for _, res := range rlist {
		if res == nil {
			continue
		}
		r := res.(map[string]string)
		if len(r) == 0 {
			continue
		}
		for obj, show := range r {
			objcontents[obj] = show
		}
	}

	for _, obj := range wobjs {
		_, ok := objcontents[obj]
		if !ok {
			err := fmt.Errorf("no defintion found for: `%s`", obj)
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
	}

	err = changeReceiver(tmpdir, Freceiver, objcontents)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	// change the objects
	// can be restored with qtechng version rebuild
	fn2 := func(n int) (interface{}, error) {
		obj := wobjs[n]
		version := Fversion
		oldsource := oldsources[obj]
		newsource := Freceiver
		err := changeEdit(version, obj, oldsource, newsource)
		return obj, err
	}
	_, errorlist := qparallel.NMap(len(wobjs), -1, fn2)
	errs := make([]error, 0)

	for _, err := range errorlist {
		if err == nil {
			continue
		}
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		Fmsg = qreport.Report(nil, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	argums = []string{"file", "ci", "--recurse", "--uid=" + FUID}
	_, serr, err = qutil.QtechNG(argums, []string{"$..ERROR"}, Fyaml, tmpdir)
	if serr != "" {
		err = fmt.Errorf("checkin of relevant sources gives error: `%s`", serr)
	}
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if len(wobjs) == 0 {
		Fmsg = qreport.Report(wobjs, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	return nil
}

func changeReceiver(tmpdir string, receiver string, objcontents map[string]string) (err error) {
	parts := strings.SplitN(receiver, "/", -1)
	parts[0] = tmpdir
	fname := filepath.Join(parts...)
	data, err := qfs.Fetch(fname)
	blob := qutil.About(data)
	if err != nil {
		return err
	}
	ext := path.Ext(fname)

	var objfile qobject.OFile
	switch ext {
	case ".d":
		objfile = new(qofile.DFile)
	case ".i":
		objfile = new(qofile.IFile)
	case ".l":
		objfile = new(qofile.LFile)
	}
	objfile.SetEditFile(receiver)
	objfile.SetRelease("0.00")
	err = qobject.Loads(objfile, blob, true)
	if err != nil {
		return err
	}
	objectlist := objfile.Objects()
	buffer := bytes.NewBuffer([]byte(objfile.Comment()))

	for _, obj := range objectlist {
		buffer.WriteString("\n\n")
		buffer.WriteString(obj.Format())
	}

	for _, obj := range objcontents {
		buffer.WriteString("\n\n")
		buffer.WriteString(obj)
	}
	buffer.WriteString("\n")
	err = qfs.Store(fname, buffer.Bytes(), "qtech")
	return err
}

func changeEdit(version string, obj string, oldsource string, newsource string) (err error) {
	release, err := qserver.Release{}.New(version, false)
	if err != nil {
		return err
	}
	fs, place := release.ObjectPlace(obj)

	body, e := fs.ReadFile(place)
	if e != nil {
		return e
	}
	if len(body) == 0 {
		return fmt.Errorf("`%s` contains no data", obj)
	}
	m := make(map[string]interface{})
	e = json.Unmarshal(body, &m)
	if e != nil {
		return nil
	}
	m["source"] = newsource
	data, _ := json.Marshal(m)
	fs.Store(place, data, "")
	err = qobject.UnLink(version, oldsource, obj)
	if err != nil {
		return err
	}
	err = qobject.Link(version, newsource, obj)
	return err
}

func extractDefinition(tmpdir string, qpath string, objs []string) (content map[string]string) {
	parts := strings.SplitN(qpath, "/", -1)
	parts[0] = tmpdir
	fname := filepath.Join(parts...)
	data, err := qfs.Fetch(fname)
	blob := qutil.About(data)
	if err != nil {
		return nil
	}
	ext := path.Ext(fname)

	var objfile qobject.OFile
	switch ext {
	case ".d":
		objfile = new(qofile.DFile)
	case ".i":
		objfile = new(qofile.IFile)
	case ".l":
		objfile = new(qofile.LFile)
	}
	objfile.SetEditFile(qpath)
	objfile.SetRelease("0.00")
	err = qobject.Loads(objfile, blob, true)
	if err != nil {
		return
	}
	content = make(map[string]string)
	objectlist := objfile.Objects()
	mapnew := make(map[string]int)
	for i, obj := range objectlist {
		mapnew[obj.String()] = i + 1
	}
	for _, obj := range objs {
		i := mapnew[obj]
		if i < 1 {
			continue
		}
		content[obj] = objectlist[i-1].Format()
	}

	buffer := bytes.NewBuffer([]byte(objfile.Comment()))

	for _, obj := range objectlist {
		name := obj.String()
		_, ok := content[name]
		if ok {
			continue
		}
		buffer.WriteString("\n\n")
		buffer.WriteString(obj.Format())
	}
	buffer.WriteString("\n")
	err = qfs.Store(fname, buffer.Bytes(), "qtech")
	if err != nil {
		return nil
	}
	return
}
