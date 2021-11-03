package action

import (
	"fmt"
	"io"
	"log"
	"strings"

	qmupip "brocade.be/goyo/lib/mupip"
	qliner "github.com/peterh/liner"
)

func Load(text string) string {
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
			help := `mupip load
		[-BE[GIN]=integer -E[ND]=integer
		-FI[LLFACTOR]=integer
		-FO[RMAT]={GO|B[INARY]|Z[WR]]}
		-I[GNORECHSET]
		-O[NERROR]={STOP|PROCEED|INTERACTIVE}
		-S[TDIN]] file-name`
			fmt.Println(help)
		case "":
		default:
			ask = false
			argums := append([]string{"LOAD"}, strings.Fields(text)...)
			stdout, stderr, err := qmupip.MUPIP(argums, "")
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
				history = "load " + text
			}
		}
		if !ask {
			break
		}
		textn, err := line.PromptWithSuggestion("load ", text, -1)
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
