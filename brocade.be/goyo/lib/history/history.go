package history

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	qutil "brocade.be/goyo/lib/util"
	qliner "github.com/peterh/liner"
)

func historyfile() string {
	dir := ""
	dirname, err := os.UserHomeDir()
	if err == nil && dirname != "" {
		dir = dirname
	}
	if dir == "" {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, ".config", "goyo")
	os.MkdirAll(dir, os.ModePerm)
	return filepath.Join(dir, "history")
}

func LoadHistory(line *qliner.State) {
	if f, err := os.Open(historyfile()); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
}

func SaveHistory(line *qliner.State) {
	fname := historyfile()
	if f, err := os.Create(fname); err != nil {
		log.Print("Error writing history file: ", err)
		return
	} else {
		line.WriteHistory(f)
		f.Close()
	}
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	f.Close()
	maxlen := 1000
	var lins []string
	last := ""
	for _, action := range lines {
		if action == last {
			continue
		}
		last = action
		key, _ := qutil.KeyText(action)
		key = strings.ToLower(key)
		if key == "bye" || key == "exit" || key == "quit" {
			continue
		}
		lins = append(lins, action)
	}
	i := 0
	if len(lins) > maxlen {

		i = len(lins) - maxlen
	}
	os.WriteFile(fname, []byte(strings.Join(lins[i:], "\n")), os.ModePerm)
}
