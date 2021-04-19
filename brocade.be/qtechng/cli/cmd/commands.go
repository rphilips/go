package cmd

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Available commands",
	Long: `
Displays a list of all available qtechng commands`,
	Example: "  qtechng commands",
	RunE:    commands}

func init() {

	rootCmd.AddCommand(commandsCmd)
}

func commands(cmd *cobra.Command, args []string) error {
	fmt.Println("Available Commands:")
	for _, command := range rootCmd.Commands() {
		use := command.Use
		pad := strings.Repeat(" ", 50-utf8.RuneCountInString(use))
		fmt.Println("  " + use + pad + command.Short)
		for _, subCommand := range command.Commands() {
			subUse := subCommand.Use
			subPad := strings.Repeat(" ", 50-utf8.RuneCountInString(subUse))
			fmt.Println("    " + subUse + subPad + subCommand.Short)
		}
	}
	return nil
}
