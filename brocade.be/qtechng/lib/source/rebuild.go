package source

import (
	qmeta "brocade.be/qtechng/lib/meta"
	qserver "brocade.be/qtechng/lib/server"
)

// Rebuild a list of qpaths

func Rebuild(batchid string, version string, qpaths []string) (err error) {
	if len(qpaths) == 0 {
		return nil
	}
	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return err
	}

	fmeta := func(qp string) qmeta.Meta { return qmeta.Meta{} }
	fdata := func(qp string) ([]byte, error) {
		fs, qpath := release.SourcePlace(qp)
		return fs.ReadFile(qpath)
	}
	_, err = StoreList(batchid, version, qpaths, true, fmeta, fdata, false)
	return err
}
