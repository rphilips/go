package cmd

import (
	"fmt"
	"log"
	"os"

	qregistry "brocade.be/base/registry"
	qssh "brocade.be/base/ssh"
	qclient "brocade.be/qtechng/lib/client"
	"github.com/spf13/cobra"
)

func preSSH(cmd *cobra.Command) {
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
		fmt.Print(catchOut)
	}
	if catchErr.Len() != 0 {
		fmt.Print(catchErr)
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
}
