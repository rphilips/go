package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
)

var fsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a binary",
	Long: `This command installs files as binaries in an appropriate way.
These binaries are always placed in the directory pointed to by the
'bindir' registry value.

The '--target' flag can give an alternative name.
The '--suid' flag can give 'suid' permissions.`,
	Args:    cobra.MaximumNArgs(1),
	Example: "qtechng fs install iiiftool-linux-amd64 --target=iiiftool --suid",
	RunE:    fsInstall,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Ftarget = ""
var Fsuid = false

func init() {
	fsInstallCmd.Flags().StringVar(&Ftarget, "target", "", "alternative name for binary")
	fsInstallCmd.Flags().BoolVar(&Fsuid, "suid", false, "a suid binary")
	fsCmd.AddCommand(fsInstallCmd)
}

func fsInstall(cmd *cobra.Command, args []string) error {
	if qregistry.Registry["error"] != "" {
		Fmsg = qreport.Report(nil, errors.New(qregistry.Registry["error"]), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	bindir := qregistry.Registry["bindir"]
	if bindir == "" || !qfs.IsDir(bindir) {
		Fmsg = qreport.Report(nil, errors.New("`bindir` is not defined or does not exist"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	newexe := qutil.AbsPath(args[0], Fcwd)
	if !qfs.IsFile(newexe) {
		Fmsg = qreport.Report(nil, errors.New("`"+newexe+"` does not exist"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	name := filepath.Base(Ftarget)
	if name == "" {
		name = filepath.Base(newexe)
	}
	name = strings.TrimSuffix(name, ".exe")
	if runtime.GOOS == "windows" {
		name += ".exe"
		Fsuid = false
	}
	oldexe := filepath.Join(bindir, name)

	if qfs.SameFile(newexe, oldexe) {
		Fmsg = qreport.Report(nil, errors.New("do not replace a binary with itself"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil

	}

	os.Remove(oldexe + ".bak")

	ftmp, err := qfs.TempFile(filepath.Dir(oldexe), "exe-")
	if err != nil {
		return err
	}
	original := newexe
	err = os.Rename(newexe, ftmp)
	if err != nil {
		err = qfs.CopyFile(newexe, ftmp, "", false)
	}
	if err != nil {
		return err
	}
	newexe = ftmp
	for i := 0; i < 2; i++ {
		os.Rename(oldexe, oldexe+".bak")
		os.Remove(oldexe)
		err = os.Rename(newexe, oldexe)
		if err == nil {
			qfs.CopyFile(oldexe, original, "", true)
		}

		if err == nil || i == 1 {
			break
		}
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		err = qfs.SetPathmode(oldexe, "scriptfile")
		if err != nil {
			return err
		}
		perm := qfs.CalcPerm("rwxrwxr-x")
		if Fsuid {
			err = os.Chmod(oldexe, perm|os.ModeSetuid|os.ModeSetgid)
		} else {
			err = os.Chmod(oldexe, perm)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
