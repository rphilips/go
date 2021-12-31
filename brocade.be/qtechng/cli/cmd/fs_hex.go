package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsHexCmd = &cobra.Command{
	Use:   "hex",
	Short: "Hexlify the arguments",
	Long: `Shows the arguments hexlified
`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs hex 'Hello World'
qtechng fs hex`,
	RunE: fsHex,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsHexCmd.Flags().BoolVar(&Ftolower, "tolower", false, "Lowercase the hex characters")
	fsCmd.AddCommand(fsHexCmd)
}

func fsHex(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	if len(args) == 0 {
		for {
			fmt.Print("Argument: ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 0 {
			return nil
		}
	}

	msg := make(map[string]string)

	for _, arg := range args {
		x := hex.EncodeToString([]byte(arg))
		if Ftolower {
			msg[arg] = strings.ToLower(x)
		} else {
			msg[arg] = strings.ToUpper(x)
		}

	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}
