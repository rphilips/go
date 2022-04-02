package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileLoopCmd = &cobra.Command{
	Use:   "loop",
	Short: "Loop over qtechng files",
	Long: `This command is used on workstations.
It monitors the local *qtechng-work-dir* and logs the changed qpaths on stdout.
It also maintains a list in the support directory of all projects on the workstation.

It is mainly used for supporting IDEs.
`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng file loop`,
	RunE:    fileLoop,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "W",
	},
}

var Fsleep int
var Fonce bool

func init() {
	fileLoopCmd.Flags().IntVar(&Fsleep, "sleep", -1, "Sleep before restarting")
	fileLoopCmd.Flags().BoolVar(&Fonce, "once", false, "Run once")
	fileCmd.AddCommand(fileLoopCmd)
}

func fileLoop(cmd *cobra.Command, args []string) error {
	startdir := qregistry.Registry["qtechng-work-dir"]
	if len(args) == 1 {
		startdir = qutil.AbsPath(args[0], Fcwd)
	}
	if !Fonce && Fsleep < 1 {
		x := qregistry.Registry["qtechng-workstation-introspect"]
		if x != "" {
			y, err := strconv.Atoi(x)
			if err != nil {
				y = 5
			}
			Fsleep = y
		}
	}
	if !Fonce && Fsleep < 1 {
		return nil
	}

	last := time.Now().AddDate(0, 0, -1)
	list := make([]string, 0)
	for {
		supportDirs(&last, startdir, Fversion)

		if !Fonce {
			d, _ := time.ParseDuration(strconv.Itoa(Fsleep) + "s")
			time.Sleep(d)
		}
		plocfils, errlist := qclient.Find(startdir, nil, Fversion, true, nil, true, "", "", nil)
		if errlist != nil {
			fmt.Println("[]")
			if Fonce {
				break
			}
			continue
		}
		if Flist != "" {
			for _, plocfil := range plocfils {
				list = append(list, plocfil.QPath)
			}
			qutil.EditList(Flist, false, list)
		}

		if plocfils == nil {
			fmt.Println("[]")
			if Fonce {
				break
			}
			continue
		}
		if len(plocfils) == 0 {
			fmt.Println("[]")
			if Fonce {
				break
			}
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
		if Fonce {
			break
		}
		continue
	}
	return nil
}

func supportDirs(last *time.Time, startdir string, version string) {
	d := time.Since(*last)
	if d.Hours() < 1 {
		return
	}
	matches, err := qfs.Find(startdir, []string{".qtechng"}, true, true, false)
	if err != nil {
		return
	}
	qpaths := make(map[string]bool)
	for _, f := range matches {
		dirname := filepath.Dir(f)
		_, qdir := dirProps(dirname)
		if qdir != "" {
			qpaths[qdir] = true
		}
	}

	result := make([]string, len(qpaths))
	i := 0
	for qdir := range qpaths {
		result[i] = qdir
		i++
	}
	sort.Strings(result)

	b, _ := json.Marshal(result)
	target := filepath.Join(qregistry.Registry["qtechng-support-dir"], "qdirs_list.json")
	qfs.Store(target, b, "qtech")
}
