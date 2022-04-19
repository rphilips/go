package status

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Status struct {
	Year   int               `json:"year"`
	Week   int               `json:"week"`
	Pcode  string            `json:"pcode"`
	Images map[string]string `json:"images"`
}

func DirStatus(dir string) (pstatus *Status, err error) {
	pstatus = new(Status)
	err = pstatus.Load(dir)
	return
}

func (pstatus *Status) Init(dir string, year int, week int) {
	if dir != "" {
		pstatus.Load(dir)
	}
	if year != 0 {
		pstatus.Year = year
	}
	if week != 0 {
		pstatus.Week = week
	}
	if dir != "" {
		pstatus.Save(dir)
	}
}

func (pstatus *Status) Load(dir string) (err error) {
	data, err := os.ReadFile(filepath.Join(dir, ".pblad"))
	if errors.Is(err, os.ErrNotExist) {
		data = []byte("{}")
	} else {
		return err
	}
	err = json.Unmarshal(data, pstatus)
	if err != nil {
		return err
	}
	return nil
}

func (pstatus *Status) Save(dir string) (err error) {
	data, err := json.MarshalIndent(*pstatus, "", "    ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(dir, ".pblad"), data, 0644)
	return err
}
