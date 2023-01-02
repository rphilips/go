package cmd

import (
	"encoding/json"
	"fmt"

	plog "brocade.be/pbladng/lib/log"
	"github.com/spf13/cobra"
)

var markCmd = &cobra.Command{
	Use:   "mark",
	Short: "Markrmation mark `gopblad`",
	Long:  `Version and build time markrmation mark the qtechng executable.`,

	Args:    cobra.MaximumNArgs(2),
	Example: `gopblad mark`,
	RunE:    mark,
}

func init() {

	rootCmd.AddCommand(markCmd)
}

func mark(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		log, err := plog.Fetch()
		if err != nil {
			return err
		}
		data, _ := json.MarshalIndent(log, "", "    ")
		fmt.Println(string(data))
		return nil
	}

	if len(args) == 2 {
		plog.SetMark(args[0], args[1])
	}

	value, stamp := plog.GetMark(args[0])
	fmt.Printf("%s:\n    %s\n    %s\n", args[0], value, stamp)

	return nil
}
