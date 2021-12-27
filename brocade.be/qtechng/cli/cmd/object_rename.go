package cmd

import (
	"bytes"
	"fmt"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qclient "brocade.be/qtechng/lib/client"
	qdfile "brocade.be/qtechng/lib/file/dfile"
	qobject "brocade.be/qtechng/lib/object"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var objectRenameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename an object",
	Long: `This command renames an object.
The first argument is the old (existing) name.
The second argument is the new name

The different steps to replace an object with name OLD
with an object with name NEW are as follows:

    - Both NEW and OLD have to be defined in an appropriate source!
      (OLD and NEW can be in different files)
	- OLD and NEW have to be of the same type (m4, i4, l4).
	- If, after the renaming, OLD has to be deleted, it is up
	  to the developer to do so.
`,
	Args:    cobra.ExactArgs(2),
	Example: `qtechng object rename m4_Old m4_New`,
	RunE:    objectRename,
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

func init() {
	objectCmd.AddCommand(objectRenameCmd)
}

func objectRename(cmd *cobra.Command, args []string) error {
	objold := args[0]
	objnew := args[1]
	if objnew == objold {
		err := fmt.Errorf("`%s`: same identifier for both objects", objold)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if len(objold) < 4 {
		err := fmt.Errorf("`%s` cannot be an object", objold)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if len(objnew) < 4 {
		err := fmt.Errorf("`%s` cannot be an object", objnew)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	pfold := ""
	pfnew := ""
	for _, p := range []string{"m4_", "l4_", "i4_"} {
		if strings.HasPrefix(objold, p) {
			pfold = p[:2]
		}
		if strings.HasPrefix(objnew, p) {
			pfnew = p[:2]
		}
	}
	if pfold == "" {
		err := fmt.Errorf("`%s` should start with m4, l4 or i4", objold)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if pfnew == "" {
		err := fmt.Errorf("`%s` should start with m4, l4 or i4", objnew)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if pfnew != pfold {
		err := fmt.Errorf("`%s` and `%s` are of different type", objnew, objold)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	release, err := qserver.Release{}.New(Fversion, true)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	fs, placeold := release.ObjectPlace(objold)
	exists, _ := fs.Exists(placeold)
	if !exists {
		err := fmt.Errorf("`%s` is not an existing object", objold)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	fs, placenew := release.ObjectPlace(objnew)
	exists, _ = fs.Exists(placenew)
	if !exists {
		err := fmt.Errorf("`%s` is not an existing object", objnew)
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	mapdepnew, err := qobject.GetDependenciesDeep(release, objnew)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	mapdepsnew := mapdepnew[objnew]
	for _, dep := range mapdepsnew {
		if dep == objold {
			err := fmt.Errorf("`%s` depends on `%s`", objold, objnew)
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
	}

	oldsource := qobject.GetEditFile(Fversion, objold)

	mapdep, err := qobject.GetDependenciesDeep(release, objold)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	mapdeps, ok := mapdep[objold]

	if !ok {
		return nil
	}

	msources := make(map[string]bool)
	for _, dep := range mapdeps {
		if dep == objnew {
			err := fmt.Errorf("`%s` depends on `%s`", objnew, objold)
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil

		}
		if !strings.HasPrefix(dep, "/") {
			dep = qobject.GetEditFile(Fversion, dep)
		}
		if dep == "" {
			continue
		}
		msources[dep] = true
	}
	sources := make([]string, len(msources))
	i := -1
	for s := range msources {
		i++
		sources[i] = s
	}

	tmpdir, err := qfs.TempDir("", "objrename.")
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	argums := []string{"source", "co", "--version=" + Fversion, "--tree", "--nature=text"}
	argums = append(argums, sources...)

	_, serr, err := qutil.QtechNG(argums, []string{"$..ERROR"}, false, tmpdir)

	if serr != "" {
		err = fmt.Errorf("checkout of relevant sources gives error: `%s`", serr)
	}
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	plocfils, errlist := qclient.Find(tmpdir, nil, Fversion, true, nil, false, "", "", nil)

	if errlist != nil {
		Fmsg = qreport.Report(nil, errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	fn := func(n int) (result interface{}, err error) {
		plocfil := plocfils[n]
		place := plocfil.Place
		qpath := plocfil.QPath
		body, err := qfs.Fetch(place)
		if err != nil {
			return nil, err
		}
		var blob []byte
		if qpath == oldsource {
			blob, err = replaceInEdit(body, objold, objnew)
		} else {
			blob, err = replaceInFile(body, objold, objnew)
		}
		if err == nil {
			err = qfs.Store(place, blob, "qtech")
		}
		return nil, err
	}
	_, errorlist := qparallel.NMap(len(plocfils), -1, fn)

	for _, e := range errorlist {
		if e != nil {
			Fmsg = qreport.Report(nil, e, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
	}

	argums = []string{"file", "ci", "--recurse", "--uid=" + FUID}
	Fmsg, _, _ = qutil.QtechNG(argums, Fjq, Fyaml, tmpdir)
	return nil
}

func replaceInEdit(body []byte, objold string, objnew string) (blob []byte, err error) {
	ty := objold[:2]
	if ty == "l4" {
		return body, nil
	}
	if ty == "i4" {
		return replaceInFile(body, objold, objnew)
	}

	blob = qutil.About(body)
	df := new(qdfile.DFile)
	err = qobject.Loads(df, blob, true)
	if err != nil {
		return
	}

	macros := df.Macros
	buffer := bytes.NewBuffer([]byte(df.Preamble))

	for _, macro := range macros {

		if macro.String() == objold {
			buffer.WriteString("\n\n")
			buffer.WriteString(macro.Format())
			continue
		}
		x := []byte(macro.Format())
		b, err := replaceInFile(x, objold, objnew)
		if err != nil {
			return nil, err
		}
		buffer.WriteString("\n\n")
		buffer.Write(b)
	}
	buffer.WriteString("\n")

	return buffer.Bytes(), nil
}

func replaceInFile(body []byte, objold string, objnew string) (blob []byte, err error) {

	ty := objold[:2]
	isobj := true
	parts := qutil.ObjectSplitter(body)
	bnew := []byte(objnew)
	for _, part := range parts {
		isobj = !isobj
		spart := string(part)
		if !isobj && !strings.Contains(spart, objold) {
			blob = append(blob, part...)
			continue
		}
		if !isobj {
			x, _ := replaceInFile(part, objold, objnew)
			blob = append(blob, x...)
			continue
		}

		if spart == objold {
			blob = append(blob, bnew...)
			continue
		}
		if ty != "l4" {
			blob = append(blob, part...)
			continue
		}
		canon, _ := qutil.DeNEDFU(spart)
		if canon != objold {
			blob = append(blob, part...)
			continue
		}
		p1 := strings.SplitN(objnew, "_", 2)
		p2 := strings.SplitN(spart, "_", 3)
		if len(p2) < 3 {
			blob = append(blob, part...)
			continue
		}
		p2[2] = p1[1]
		part = []byte(strings.Join(p2, "_"))
		blob = append(blob, part...)
	}
	return
}
