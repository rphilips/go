package cmd

import (
	"io"
	"log"
	"os"
	"strings"

	qsed "github.com/rwtodd/Go.Sed/sed"
	"github.com/spf13/cobra"
)

var fileSedCmd = &cobra.Command{
	Use:     "sed",
	Short:   "Sed stream editor",
	Long:    `Filters stdin through a sed statement and writes on stdout`,
	Example: "  qtechng file sed '/Hello/d'",
	Args:    cobra.ExactArgs(1),
	RunE:    fileSed,
}

func init() {
	fileCmd.AddCommand(fileSedCmd)
}

func fileSed(cmd *cobra.Command, args []string) (err error) {
	program := strings.NewReader(args[0])
	engine, err := qsed.New(program)
	if err == nil {
		if Fstdout == "" || Ftransported {
			_, err = io.Copy(os.Stdout, engine.Wrap(os.Stdin))
			return err
		}

		f, err := os.Create(Fstdout)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		_, err = io.Copy(f, engine.Wrap(os.Stdin))
	}

	return err
}
