package image

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	pbstatus "brocade.be/pbladng/status"
)

// ImageMap creates a map with the identifier (lowercase) mapped to the relpath to dir
func ImageMap(dir string) map[string]string {
	m := make(map[string]string)
	fn := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		path, _ = filepath.Rel(dir, path)
		name := info.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}
		k := strings.IndexAny(name, "-_.")
		index := name[:k]
		found, ok := m[index]
		if ok {
			idirname := filepath.Join(filepath.Dir(path), "x")
			fdirname := filepath.Join(filepath.Dir(found), "x")
			if !strings.HasPrefix(fdirname, idirname) {
				return nil
			}
		}
		m[index] = path
		return nil
	}
	filepath.WalkDir(dir, fn)
	return m
}

func ImageRef(images []string, dir string) (err error) {
	pstatus, err := pbstatus.DirStatus(dir)
	if err != nil {
		return err
	}
	refimages := make(map[string]string)
	m := ImageMap(dir)

	notfound := make(map[string]bool)
	toomany := make(map[string]bool)
	suffix := strconv.Itoa(pstatus.Week)
	if len(suffix) == 1 {
		suffix = "0" + suffix
	}
	suffix += ".jpg"
	prefix := "F" + pstatus.Pcode
	for _, imag := range images {
		k := strings.IndexAny(imag, "-_.")
		index := imag[:k]
		if m[index] == "" {
			notfound[imag] = true
		}
		if len(notfound) != 0 {
			continue
		}
		i := len(refimages)
		if i > 25 {
			toomany[imag] = true
			continue
		}
		ch := string(rune(97 + i))
		refimages[index] = prefix + ch + suffix
	}
	pstatus.Images = refimages
	err = pstatus.Save(dir)
	if err != nil {
		return err
	}
	if len(notfound) != 0 {
		nf := make([]string, len(notfound))
		i := 0
		for imag := range notfound {
			nf[i] = imag
			i++
		}
		return fmt.Errorf("ERROR images: %s not found!", strings.Join(nf, ", "))
	}
	if len(toomany) != 0 {
		nf := make([]string, len(notfound))
		i := 0
		for imag := range toomany {
			nf[i] = imag
			i++
		}
		return fmt.Errorf("ERROR images: %s too many!", strings.Join(nf, ", "))
	}

	return err

}
