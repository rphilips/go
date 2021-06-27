package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileLoopCmd = &cobra.Command{
	Use:     "loop",
	Short:   "Loop QtechNG files",
	Long:    `Command `,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng file loop`,
	RunE:    fileLoop,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

var Fsleep int

func init() {
	fileLoopCmd.Flags().IntVar(&Fsleep, "sleep", -1, "Sleep before restarting")
	fileCmd.AddCommand(fileLoopCmd)
}

func fileLoop(cmd *cobra.Command, args []string) error {
	startdir := qregistry.Registry["qtechng-work-dir"]
	if len(args) == 1 {
		startdir = qutil.AbsPath(args[0], Fcwd)
	}
	if Fsleep < 1 {
		x := qregistry.Registry["qtechng-workstation-introspect"]
		if x != "" {
			y, err := strconv.Atoi(x)
			if err != nil {
				y = 5
			}
			Fsleep = y
		}
	}
	if Fsleep < 1 {
		return nil
	}

	for {
		d, _ := time.ParseDuration(strconv.Itoa(Fsleep) + "s")
		time.Sleep(d)
		plocfils, errlist := qclient.Find(startdir, nil, Fversion, true, nil, true)
		if errlist != nil {
			fmt.Println("[]")
			continue
		}
		if plocfils == nil {
			fmt.Println("[]")
			continue
		}
		if len(plocfils) == 0 {
			fmt.Println("[]")
			continue
		}
		tip := make([]string, 0, len(plocfils))
		for _, plocfil := range plocfils {
			if plocfil == nil {
				continue
			}
			tip = append(tip, plocfil.Place)
		}
		b, _ := json.Marshal(tip)
		fmt.Println(string(b))
		continue
	}

}
