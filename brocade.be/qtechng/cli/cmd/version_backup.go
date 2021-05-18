package cmd

import (
	"archive/tar"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

var versionBackupCmd = &cobra.Command{
	Use:   "backup version",
	Short: "Backup of version",
	Long: `Backup is in tar (PAX) format. Meta data is attached as well
	The result is always brocade-version.tar in the current directory.

This file is usable with tar but it contains the QtechNG meta data in 
PAX extended header records	with the namespace BROCADE`,
	Args:    cobra.ExactArgs(1),
	Example: "qtechng version backup 0.00",
	RunE:    versionBackup,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	versionCmd.AddCommand(versionBackupCmd)
}

func versionBackup(cmd *cobra.Command, args []string) error {
	h := time.Now()
	t := h.Format(time.RFC3339)[:19]
	t = strings.ReplaceAll(t, ":", "")
	t = strings.ReplaceAll(t, "-", "")
	r := qserver.Canon(args[0])
	release, _ := qserver.Release{}.New(r, true)
	ok, _ := release.Exists("/source/data")
	if !ok {
		err := &qerror.QError{
			Ref: []string{"backup.notexist"},
			Msg: []string{"Vversion does not exist."},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}

	fs := release.FS()
	qpaths := fs.Glob("/", nil, false)

	// tar
	errlist := make([]error, 0)
	tarfile := qutil.AbsPath("brocade-"+r+"-"+t+".tar", Fcwd)
	ftar, err := os.Create(tarfile)

	if err != nil {
		return err
	}
	defer ftar.Close()

	tw := tar.NewWriter(ftar)

	for _, qpath := range qpaths {
		x, _ := fs.RealPath(qpath)
		fi, err := os.Stat(x)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		source, err := qsource.Source{}.New(r, qpath, true)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		content, err := source.Fetch()
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		meta, err := qmeta.Meta{}.New(r, qpath)
		if err != nil {
			errlist = append(errlist, err)
			continue
		}
		pax := map[string]string{
			"BROCADE.cu": meta.Cu,
			"BROCADE.mu": meta.Mu,
			"BROCADE.ct": meta.Ct,
			"BROCADE.mt": meta.Mt,
			"BROCADE.it": meta.It,
			"BROCADE.ft": meta.Ft,
		}

		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"version.backup.fiheader"},
				Version: r,
				QPath:   qpath,
				Msg:     []string{err.Error()},
			}
			err = qerror.QErrorTune(err, e)
			errlist = append(errlist, err)
			continue
		}
		hdr.Name = qpath[1:]
		hdr.Format = tar.FormatPAX
		hdr.PAXRecords = pax
		if err := tw.WriteHeader(hdr); err != nil {
			e := &qerror.QError{
				Ref:     []string{"version.backup.header"},
				Version: r,
				QPath:   qpath,
				Msg:     []string{err.Error()},
			}
			err = qerror.QErrorTune(err, e)
			errlist = append(errlist, err)
			continue
		}
		if _, err := tw.Write(content); err != nil {
			e := &qerror.QError{
				Ref:     []string{"version.backup.body"},
				Version: r,
				QPath:   qpath,
				Msg:     []string{err.Error()},
			}
			err = qerror.QErrorTune(err, e)
			errlist = append(errlist, err)
			continue
		}
	}
	tw.Flush()
	tw.Close()

	msg := make(map[string]string)
	msg["status"] = "Backup FAILED"
	if len(errlist) == 0 {
		errlist = nil
		msg["status"] = "Backup SUCCESS to `" + tarfile + "`"
	}
	Fmsg = qreport.Report(msg, errlist, Fjq, Fyaml)
	return nil
}
