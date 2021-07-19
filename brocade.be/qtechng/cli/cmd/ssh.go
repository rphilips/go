package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

func preSSH(cmd *cobra.Command, catch func(s string) string) {
	if !Fremote {
		return
	}
	payload := qclient.Payload{
		ID:     "Once",
		UID:    FUID,
		CMD:    "qtechng",
		Origin: QtechType,
		Args:   os.Args[1:],
	}
	whowhere := FUID + "@" + qregistry.Registry["qtechng-server"]
	catchOut, catchErr, err := qssh.SSHcmd(&payload, whowhere)
	if err != nil {
		log.Fatal("cmd/ssh/1:\n", err, "\n====\n", catchErr)
	}
	if catchOut.Len() != 0 {
		b := catchOut.Bytes()
		if Flist != "" && len(b) != 0 && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'W') && qregistry.Registry["qtechng-support-dir"] != "" {
			qutil.FillList(Flist, b)
		}
		if catch == nil {
			fmt.Print(catchOut)
		}
	}
	if catchErr.Len() != 0 {
		fmt.Print(catchErr)
	}
	if catch != nil {
		fmt.Print(catch(catchOut.String()))
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
}

func ReadSSHAll(fname string) ([]byte, error) {
	payload := qclient.Payload{
		ID:     "Once",
		UID:    FUID,
		CMD:    "qtechng",
		Origin: QtechType,
		Args:   []string{"fs", "cat", fname},
	}
	whowhere := FUID + "@" + qregistry.Registry["qtechng-server"]
	catchOut, catchErr, err := qssh.SSHcmd(&payload, whowhere)
	if err != nil {
		log.Fatal("cmd/cat/1:\n", err, "\n====\n", catchErr)
	}
	return catchOut.Bytes(), nil
}
