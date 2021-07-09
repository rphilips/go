package cmd

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var systemSleepCmd = &cobra.Command{
	Use:     "sleep",
	Short:   "Sleep a number of seconds",
	Long:    `Sleep a number of seconds`,
	Args:    cobra.MaximumNArgs(1),
	Example: `  qtechng system sleep`,
	RunE:    systemSleep,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	systemCmd.AddCommand(systemSleepCmd)
}

func systemSleep(cmd *cobra.Command, args []string) error {

	if len(args) == 0 {
		return nil
	}
	x := args[0]
	y := 0
	if x != "" {
		var err error
		y, err = strconv.Atoi(x)
		if err != nil {
			return err
		}
	}
	if y < 1 {
		return nil
	}
	d, _ := time.ParseDuration(strconv.Itoa(y) + "s")
	time.Sleep(d)
	return nil
}
