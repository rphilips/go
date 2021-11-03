package action

import (
	"fmt"
	"io"
	"log"
	"strings"

	qmupip "brocade.be/goyo/lib/mupip"
	qliner "github.com/peterh/liner"
)

func Extract(text string) string {
	ask := true
	history := ""
	line := qliner.NewLiner()
	line.SetCtrlCAborts(true)
	defer line.Close()
	for {
		if !ask {
			break
		}
		switch text {
		case "?":
			help := `mupip extract
			[-FO[RMAT]={GO|B[INARY]|Z[WR]}
			-FR[EEZE]
			-LA[BEL]=text
			-[NO]L[OG]
			-[NO]NULL_IV
			-R[EGION]=region-list
			-SE[LECT]=global-name-list]
			]
			{-ST[DOUT]|file-name}`
			fmt.Println(help)
		case "":
		default:
			ask = false
			args := []string{"EXTRACT", "-FO=Z"}
			argums := strings.Fields(text)
			if len(argums) > 1 {
				glo := strings.TrimPrefix(argums[0], "^")
				args = append(args, "-SE="+glo, "-ST="+argums[1])
			}
			if len(argums) != 0 {
				glo := strings.TrimPrefix(argums[0], "^")
				args = append(args, "-SE="+glo, glo+".zwr")
			}
			if len(argums) == 0 {
				continue
			}
			ask = false
			stdout, stderr, err := qmupip.MUPIP(args, "")
			if strings.TrimSpace(stderr) != "" {
				fmt.Println(stderr)
			}
			if strings.TrimSpace(stdout) != "" {
				fmt.Println(stdout)
			}
			if err != nil {
				fmt.Println(err)
				ask = true
			} else {
				history = "extract " + text
			}
		}
		if !ask {
			break
		}
		textn, err := line.PromptWithSuggestion("extract ", text, -1)
		if err == qliner.ErrPromptAborted || textn == "" || err == io.EOF {
			text = ""
			break
		}
		if err != nil {
			log.Print("Error reading line: ", err)
			continue
		}
		text = textn
		continue
	}
	return history

}
