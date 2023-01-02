package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	bstring "brocade.be/base/strings"
)

var doubleCmd = &cobra.Command{
	Use:   "double",
	Short: "import documents from the correspondents directories",
	Long:  "import documents from the correspondents directories",

	Args:    cobra.NoArgs,
	Example: `gopblad double`,
	RunE:    double,
}

func init() {
	rootCmd.AddCommand(doubleCmd)
}

func double(cmd *cobra.Command, args []string) error {

	doubles, err := bfs.Doubles(Fcwd)

	if err != nil {
		return fmt.Errorf("looking for doubles: %s", err)
	}

	if len(doubles) != 0 {
		fmt.Println(doubles)
		fmt.Println("Found doubles in", Fcwd+":\n")
		for _, d := range doubles {
			fmt.Println(bstring.JSON(d))
			fmt.Println()
		}
	}

	return err
}
