package image

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	pfs "brocade.be/base/fs"
	"brocade.be/pbladng/registry"
	pregistry "brocade.be/pbladng/registry"
	pbstatus "brocade.be/pbladng/status"
	ptools "brocade.be/pbladng/tools"
)

type Image struct {
	Name      string
	Legend    string
	Copyright string
	Fname     string
	Lineno    int
}

func New(line ptools.Line, dir string) (image Image, err error) {
	s := line.L
	lineno := line.NR
	name, legend, ok := strings.Cut(s, ".jpg")
	if !ok {
		err = ptools.Error("image-jpg", lineno, "line should contain .jpg")
		return
	}
	name += "."
	k := strings.IndexAny(name, "-_.")
	name = name[:k]
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		err = ptools.Error("topic-image-name", lineno, "name is empty")
		return
	}
	if dir == "" {
		dir = pregistry.Registry["workspace-path"]
	}
	imgmap := ImageMap(dir)
	if imgmap[name] == "" {
		err = ptools.Error("topic-image-file", lineno, "cannot find image")
		return
	}
	legend = strings.TrimSpace(legend)

	legend, copyright, ok := strings.Cut(s, "©")
	if !ok {
		legend, copyright, ok = strings.Cut(s, "copyright")
	}
	if !ok {
		legend, copyright, ok = strings.Cut(s, "Copyright")
	}
	legend = strings.TrimSpace(legend)
	copyright = strings.TrimSpace(copyright)

	image = Image{
		Name:      name,
		Legend:    legend,
		Copyright: copyright,
		Fname:     imgmap[name],
		Lineno:    lineno,
	}
	return
}

// ImageMap creates a map with the identifier (lowercase) mapped to the relpath to dir
func ImageMap(dir string) map[string]string {
	if dir == "" {
		dir = pregistry.Registry["workspace-path"]
	}
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
	if dir == "" {
		dir = pregistry.Registry["workspace-path"]
	}
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

func ReduceSize(imgpath string, kbsize int) (err error) {
	if kbsize < 0 {
		kbsize, _ = strconv.Atoi(pregistry.Registry["image-size-kb"])
	}

	fi, err := os.Stat(imgpath)
	if err != nil {
		return
	}
	if int64(kbsize)*1024 > fi.Size() {
		return
	}

	small := strings.TrimSuffix(imgpath, filepath.Ext(imgpath)) + "__small__.jpg"
	pfs.Rmpath(small)
	worker := registry.Registry["image-exe"]
	if worker == "" {
		worker = "convert"
	}
	binary, err := exec.LookPath(worker)
	if err != nil {
		return
	}
	args := []string{imgpath, "-define", "jpeg:extent=" + strconv.Itoa(kbsize) + "kb", small}
	out, err := exec.Command(binary, args...).Output()
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), string(out))
		return
	}
	fi, err = os.Stat(small)
	if err != nil {
		return
	}
	pfs.Rmpath(imgpath + ".ori")
	err = os.Rename(imgpath, imgpath+".ori")
	if err != nil {
		return
	}
	err = os.Rename(small, imgpath)
	return
}

func ChangeType(imgpath string) (err error) {

	_, err = os.Stat(imgpath)
	if err != nil {
		return
	}
	ext := strings.ToLower(filepath.Ext(imgpath))
	if ext != ".png" {
		err = fmt.Errorf("not the right extension: `%s`", ext)
		return
	}

	jpeg := strings.TrimSuffix(imgpath, filepath.Ext(imgpath)) + "__jpeg__.jpg"
	pfs.Rmpath()
	worker := registry.Registry["image-exe"]
	if worker == "" {
		worker = "convert"
	}
	binary, err := exec.LookPath(worker)
	if err != nil {
		return
	}
	args := []string{imgpath, "-define", "jpeg:extent=" + strconv.Itoa(kbsize) + "kb", jpeg}
	out, err := exec.Command(binary, args...).Output()
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), string(out))
		return
	}
	fi, err = os.Stat(jpeg)
	if err != nil {
		return
	}
	pfs.Rmpath(imgpath + ".ori")
	err = os.Rename(imgpath, imgpath+".ori")
	if err != nil {
		return
	}
	err = os.Rename(jpeg, imgpath)
	return
}
