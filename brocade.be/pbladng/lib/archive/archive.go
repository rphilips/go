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

func FindBefore(id string, mailed bool) (place string, m *pmanuscript.Manuscript, err error) {
	if id == "" {
		now := time.Now()
		year := now.Year()
		id = strconv.Itoa(year)
	}
	if !strings.Contains(id, "-") {
		id = id + "-99"
	}
	syear, sweek, _ := strings.Cut(id, "-")
	year, err := strconv.Atoi(syear)
	if err != nil {
		return
	}
	_, err = strconv.Atoi(sweek)
	if err != nil {
		return
	}
	for {
		if year < 2005 {
			return "", nil, fmt.Errorf("no manuscripts found")
		}
		dir := filepath.Join(arcdir, strconv.Itoa(year))
		files, err := os.ReadDir(dir)
		if err != nil {
			year--
			sweek = "99"
			continue
		}
		weeks := make([]string, 0)
		for _, w := range files {
			name := w.Name()
			base := filepath.Base(name)
			if len(name) != 2 {
				continue
			}
			if strings.TrimLeft(name, "1234567890") != "" {
				continue
			}
			if base < sweek {
				weeks = append(weeks, base)
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(weeks)))
		for _, week := range weeks {
			if week >= sweek {
				continue
			}
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
		year--
		sweek = "99"
	}

}
