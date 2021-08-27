package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qreport "brocade.be/qtechng/lib/report"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileCiCmd = &cobra.Command{
	Use:   "ci",
	Short: "Check in qtechng files",
	Long:  `Stores local files in the qtechng repository` + Mfiles,
	Args:  cobra.MinimumNArgs(0),
	Example: `qtechng file ci application/bcawedit.m install.py cwd=../catalografie
qtechng file ci`,
	RunE:   fileCi,
	PreRun: preCi,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

func init() {
	fileCiCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileCiCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileCiCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileCiCmd)
}

func fileCi(cmd *cobra.Command, args []string) error {

	if strings.ContainsRune(QtechType, 'B') {
		Fcargo = storeRep(Fpayload)
		if Ftransported {
			qclient.SendCargo(Fcargo)
			return nil
		}
	}

	type lister struct {
		Release string `json:"version"`
		QPath   string `json:"qpath"`
		File    string `json:"file"`
		Url     string `json:"fileurl"`
		Changed bool   `json:"changed"`
		Time    string `json:"time"`
		Digest  string `json:"digest"`
		Cu      string `json:"cu"`
		Mu      string `json:"mu"`
		Ct      string `json:"ct"`
		Mt      string `json:"mt"`
	}

	result := make([]lister, 0)

	dirs := make(map[string][]*qclient.LocalFile)

	for _, tr := range Fcargo.Transports {
		locfil := tr.LocFile
		place := locfil.Place
		dir := filepath.Dir(place)
		_, ok := dirs[dir]
		if !ok {
			dirs[dir] = make([]*qclient.LocalFile, 0)
		}
		dirs[dir] = append(dirs[dir], &locfil)
	}

	for dir, plocfiles := range dirs {
		locfiles := make([]qclient.LocalFile, 0)
		for _, plocfil := range plocfiles {
			place := plocfil.Place
			mt, e := qfs.GetMTime(place)
			if e != nil {
				continue
			}
			t := mt.Format(time.RFC3339)
			plocfil.Time = t
			locfiles = append(locfiles, *plocfil)
			result = append(result, lister{
				Release: plocfil.Release,
				QPath:   plocfil.QPath,
				File:    place,
				Url:     qutil.FileURL(place, "", -1),
				Time:    plocfil.Time,
				Digest:  plocfil.Digest,
				Cu:      plocfil.Cu,
				Mu:      plocfil.Mu,
				Ct:      plocfil.Ct,
				Mt:      plocfil.Mt,
			})
		}
		if len(locfiles) > 0 {
			d := new(qclient.Dir)
			d.Dir = dir
			d.Add(locfiles...)
		}
	}

	if Fmsg == "" {
		Fmsg = qreport.Report(result, Fcargo.Error, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	return nil
}

func preCi(cmd *cobra.Command, args []string) {
	if Ftransported {
		return
	}

	var errlist []error
	Fpayload, errlist = getPayload(args, FUID, Fcwd, Fversion, Frecurse, Fqpattern, Finlist, Fnotinlist)
	if len(errlist) == 0 {
		errlist = nil
	} else {
		errlist = qerror.FlattenErrors(qerror.ErrorSlice(errlist))
	}

	if errlist != nil || Fpayload == nil {
		Fmsg = qreport.Report(nil, qerror.ErrorSlice(errlist), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
		return
	}

	if !strings.ContainsRune(QtechType, 'B') {
		whowhere := qregistry.Registry["qtechng-server"]
		if !strings.Contains(whowhere, "@") {
			whowhere = qregistry.Registry["qtechng-user"] + "@" + whowhere
		}
		catchOut, catchErr, err := qssh.SSHcmd(Fpayload, whowhere)
		if err != nil {
			log.Fatal("cmd/file_ci/preCi/1:\n", err)
		}
		if catchErr.Len() != 0 {
			log.Fatal("cmd/file_ci/preCi/2:\n", catchErr)
		}
		Fcargo = qclient.ReceiveCargo(catchOut)
	}
}

func getPayload(args []string, uid string, cwd string, version string, recurse bool, patterns []string, inlist string, notinlist string) (payload *qclient.Payload, errlist []error) {
	plocfils, elist := qclient.Find(cwd, args, version, recurse, patterns, true, inlist, notinlist, nil)
	errlist = make([]error, 0)
	if elist != nil {
		errlist = append(errlist, elist)
	}
	if len(plocfils) == 0 {
		return nil, errlist
	}

	transports := make([]qclient.Transport, 0)

	for _, plocfil := range plocfils {
		place := plocfil.Place
		mt, e := qfs.GetMTime(place)
		if e != nil {
			err := qerror.QError{
				Ref:  []string{"ci.read.mtime"},
				File: place,
				Msg:  []string{"`" + place + "` error with modification time: " + e.Error()},
			}
			errlist = append(errlist, err)
			continue
		}
		touch := mt.Format(time.RFC3339)

		if plocfil.Time == touch {
			continue
		}
		blob, e := os.ReadFile(place)
		if e != nil {
			err := qerror.QError{
				Ref:  []string{"ci.read.file"},
				File: place,
				Msg:  []string{"`" + place + "` read with error: " + e.Error()},
			}
			errlist = append(errlist, err)
			continue
		}
		transports = append(transports, qclient.Transport{
			LocFile: *plocfil,
			Body:    blob,
		})
	}
	payload = &qclient.Payload{
		ID:         "Once",
		UID:        uid,
		CMD:        "qtechng",
		Origin:     QtechType,
		Args:       os.Args[1:],
		Transports: transports,
	}
	return payload, errlist
}

func storeRep(payload *qclient.Payload) (pcargo *qclient.Cargo) {

	errlist := make([]error, 0)

	pcargo = &qclient.Cargo{}

	versions := make(map[string][]string)
	mpaths := make(map[string]int)
	h := time.Now().Local()
	t := h.Format(time.RFC3339)
	uid := ""

	plocfils := make([]*qclient.LocalFile, len(payload.Transports))
	for i := range payload.Transports {
		plocfils[i] = &(payload.Transports[i].LocFile)
		uid = payload.UID
	}

	for i, locfil := range plocfils {
		r := locfil.Release
		_, ok := versions[r]
		if !ok {
			versions[r] = make([]string, 0)
		}
		qpath := locfil.QPath
		ipath := r + " " + qpath
		versions[r] = append(versions[r], qpath)
		mpaths[ipath] = i
	}

	stored := make([]qclient.Transport, 0)
	for version, qpaths := range versions {

		fmeta := func(qpath string) qmeta.Meta {
			ipath := version + " " + qpath
			i := mpaths[ipath]
			digest := plocfils[i].Digest
			return qmeta.Meta{
				Mt:     t,
				Mu:     uid,
				Digest: digest,
			}
		}

		fdata := func(qpath string) ([]byte, error) {
			ipath := version + " " + qpath
			i := mpaths[ipath]
			blob := payload.Transports[i].Body
			return blob, nil
		}

		results, errs := qsource.StoreList("install", version, qpaths, false, fmeta, fdata, true)
		if errs != nil {
			errlist = append(errlist, errs)
		}
		for qpath, pmeta := range results {
			if pmeta == nil {
				continue
			}

			ipath := version + " " + qpath
			i := mpaths[ipath]
			locfil := plocfils[i]
			locfil.Cu = pmeta.Cu
			locfil.Mu = pmeta.Mu
			locfil.Ct = pmeta.Ct
			locfil.Mt = pmeta.Mt
			locfil.Digest = pmeta.Digest
			locfil.Time = ""
			tr := qclient.Transport{
				LocFile: *locfil,
			}
			stored = append(stored, tr)

		}
	}
	if len(errlist) == 0 {
		pcargo.Error = nil
	} else {
		pcargo.Error = qerror.ErrorSlice(errlist)
	}
	pcargo.Transports = stored
	return pcargo
}
