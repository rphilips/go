package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	qlog "brocade.be/qtechng/lib/log"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var logPanicCmd = &cobra.Command{
	Use:   "panic",
	Short: "Show panics in log files",
	Long:  `This command shows panics in the qtechng log files`,
	Args:  cobra.MaximumNArgs(1),
	Example: `qtechng log panic
qtechng log panic 2021-07-30`,
	RunE: logPanic,
}

func init() {
	logCmd.AddCommand(logPanicCmd)
}

func logPanic(cmd *cobra.Command, args []string) error {
	qlog.Pack()
	when := ""
	if len(args) == 0 {
		h := time.Now()
		when = h.Format(time.RFC3339)[:10]
	} else {
		when = args[0]
	}
	if when == "" {
		Fmsg = qreport.Report(nil, errors.New("argument should not be empty"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	info, err := qlog.Panic(when)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	if len(info) == 0 {
		Fmsg = qreport.Report("No panics found!", err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	result := strings.Join(info, "\n\n"+strings.Repeat("#", 72)+"\n\n")
	if Fstdout == "" || Ftransported {
		fmt.Print(result)
	} else {
		f, err := os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer f.Close()
		io.Copy(bytes.NewBufferString(result), f)
	}

	return nil
}
