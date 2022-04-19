package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	_ "modernc.org/sqlite"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsTempdirCmd = &cobra.Command{
	Use:   "tempdir",
	Short: "tempdir",
	Long: `Creates a temporary directory

- With 1 argument: this argument is considered a directory, the result is an empty sub-directory
- Without arguments, the software choses an appropriate directory to create the sub-directory in

The flag '--template=...' gives the prefix of the sub-directory.
You can also put a '*' in the template: the variable part of the sub-directory replaces
the '*'
`,
	Args: cobra.MaximumNArgs(1),
	Example: `qtechng fs tempdir
qtechng fs tempdir .
qtechng fs tempdir --template=mytmp
qtechng fs tempdir --template='mytmp*dir'`,
	RunE: fstempdir,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

var Ftemplate = ""

func init() {
	fsTempdirCmd.Flags().StringVar(&Ftemplate, "template", "", "Template for temporary directory")
	fsCmd.AddCommand(fsTempdirCmd)
}

func fstempdir(cmd *cobra.Command, args []string) error {
	scratchdir := qregistry.Registry["scratch-dir"]
	if len(args) != 0 {
		scratchdir = args[0]
	}
	dir, err := ioutil.TempDir(scratchdir, Ftemplate)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-tempdir-fail")
		return nil
	}

	if Fstdout != "" {
		f, err := os.Create(Fstdout)

		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-tempdir-stdout")
			return nil
		}
		w := bufio.NewWriter(f)
		fmt.Fprint(w, dir)
		w.Flush()
		f.Close()
		return nil
	}

	fi, _ := os.Stdout.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		fmt.Print(dir)
		return nil
	}
	fmt.Println(dir)

	return nil
}
