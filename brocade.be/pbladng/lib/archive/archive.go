package archive

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	pfs "brocade.be/pbladng/lib/fs"
	pmanuscript "brocade.be/pbladng/lib/manuscript"
)

var arcdir = pfs.FName(pfs.Base + "/archive/manuscripts")

func FindLast(id string, mailed bool) (string, *pmanuscript.Manuscript, error) {
	now := time.Now()
	year := 2 + now.Year()
	week = 54
	if id != "" {
		syear, sweek, _ := strings.Cut(id, "-")
		year, _ = strconv.Atoi(syear)
		if sweek == "1" {

			week, _ := strconv.Atoi(sweek)
			year++
		}
	}
	for {
		year--
		if year < 2005 {
			return "", nil, fmt.Errorf("no manuscripts found")
		}
		dir := filepath.Join(arcdir, strconv.Itoa(year))
		files, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		weeks := make([]string, 0)
		for _, week := range files {
			name := week.Name()
			base := filepath.Base(name)
			if len(name) != 2 {
				continue
			}
			if strings.TrimLeft(name, "1234567890") != "" {
				continue
			}
			weeks = append(weeks, base)
		}
		sort.Sort(sort.Reverse(sort.StringSlice(weeks)))
		for _, week := range weeks {
			fname := filepath.Join(dir, week, "week.pb")
			f, err := os.Open(fname)
			if err != nil {
				continue
			}
			source := bufio.NewReader(f)
			m, err := pmanuscript.Parse(source)
			if err != nil {
				return "", nil, fmt.Errorf("error in %s: %s", fname, err.Error())
			}
			if !mailed || m.Mailed != nil {
				return fname, m, nil
			}
		}
	}

}
