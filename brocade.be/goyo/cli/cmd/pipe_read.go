package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var pipereadCmd = &cobra.Command{
	Use:     "read",
	Short:   "set up named pipe for reading",
	Long:    `set up named pipe for reading`,
	Args:    cobra.NoArgs,
	Example: `goyo pipe read --pipe=/tmp/goyo`,
	RunE:    piperead,
}

func init() {
	pipeCmd.AddCommand(pipereadCmd)
}

func piperead(cmd *cobra.Command, args []string) error {
	if Fpipe == "" {
		log.Fatal("Missing named pipe")
	}
	if _, err := os.Stat(Fpipe); err != nil {
		syscall.Mkfifo(Fpipe, 0600)
	}
	stdout, _ := os.OpenFile(Fpipe, os.O_RDONLY, 0600)
	reader := bufio.NewReader(stdout)
	for {
		text, err := reader.ReadString('\n')
		if strings.HasSuffix(text, "<[<end>]>\n") {
			text = strings.TrimSuffix(text, "<[<end>]>\n")
			fmt.Print(text)
			break
		}
		if err == io.EOF {
			fmt.Println(text)
			break
		}
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		fmt.Print(text)
	}

	return nil
}
