package util

import (
	"os/exec"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

func RefreshBinary() (err error) {
	exe := qregistry.Registry["qtechng-exe"]
	if exe == "" {
		return
	}
	pexe, err := exec.LookPath(exe)
	if err != nil {
		return err
	}
	tmp, err := qfs.TempFile("", "qtechng-bin-")
	if err != nil {
		return err
	}
	err = qfs.GetURL(qregistry.Registry["qtechng-url"], tmp, "tempfile")
	if err != nil {
		return err
	}

	return qfs.RefreshEXE(pexe, tmp)
}
