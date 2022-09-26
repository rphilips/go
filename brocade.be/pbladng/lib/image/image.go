package image

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	pfs "brocade.be/pbladng/lib/fs"
	"brocade.be/pbladng/lib/registry"
	pregistry "brocade.be/pbladng/lib/registry"
	pbstatus "brocade.be/pbladng/lib/status"
	ptools "brocade.be/pbladng/lib/tools"
)

type Image struct {
	Name      string
	Legend    string
	Copyright string
	Fname     string
	Lineno    int
}

func New(line ptools.Line, dir string, checkextern bool, cpright string) (image Image, err error) {
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
		dir = pfs.FName("workspace")
	}
	imgmap := ImageMap(dir)
	if imgmap[name] == "" && checkextern {
		err = ptools.Error("topic-image-file", lineno, "cannot find image `"+name+"` in "+"`"+dir+"`")
		return
	}
	legend = strings.TrimSpace(legend)
	copyright := ""
	for _, decider := range []string{"©", " cr ", "copyright", "Copyright"} {
		x, y, ok := strings.Cut(" "+legend, decider)
		if !ok {
			continue
		}
		y = strings.TrimLeft(strings.TrimSpace(y), "!,:;. ")
		if y == "" {
			continue
		}
		copyright = y
		legend = strings.TrimRight(strings.TrimSpace(x), ",:;. ")
		break
	}
	if copyright == "" {
		copyright = cpright
	}

	if copyright == "" {
		err = ptools.Error("topic-image-copyright", lineno, "no copyright for `"+name+"`")
		return
	}

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
		dir = pfs.FName("workspace")
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
		dir = pregistry.Registry["workspace-dir"].(string)
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
		kbsize, _ = strconv.Atoi(pregistry.Registry["image-size-kb"].(string))
	}

	fi, err := os.Stat(imgpath)
	if err != nil {
		return
	}
	if int64(kbsize)*1024 > fi.Size() {
		return
	}

	small := strings.TrimSuffix(imgpath, filepath.Ext(imgpath)) + "__small__.jpg"
	bfs.Rmpath(small)
	worker := registry.Registry["image-resize-exe"].([]any)
	keys := map[string]string{"source": imgpath, "kbsize": strconv.Itoa(kbsize), "target": small}
	out, err := ptools.Launch(worker, keys, "", true)
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), string(out))
		return
	}
	fi, err = os.Stat(small)
	if err != nil {
		return
	}
	bfs.Rmpath(imgpath + ".ori")
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
	ext := filepath.Ext(imgpath)

	if ext == ".jpg" {
		return
	}
	lext := strings.ToLower(ext)

	if lext == ".jpg" || lext == ".jpeg" {
		jpeg := strings.TrimSuffix(imgpath, ext) + ".jpg"
		err = os.Rename(imgpath, jpeg)
		return
	}

	if lext != ".png" {
		err = fmt.Errorf("not the right extension: `%s`", ext)
		return
	}

	jpeg := strings.TrimSuffix(imgpath, ext) + "__jpeg__.jpg"
	bfs.Rmpath(jpeg)
	worker := registry.Registry["image-resize-exe"].([]any)
	keys := map[string]string{"source": imgpath, "kbsize": pregistry.Registry["image-size-kb"].(string), "target": jpeg}
	out, err := ptools.Launch(worker, keys, "", true)
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), string(out))
		return
	}
	_, err = os.Stat(jpeg)
	if err != nil {
		return
	}
	bfs.Rmpath(imgpath + ".ori")
	err = os.Rename(imgpath, imgpath+".ori")
	if err != nil {
		return
	}
	err = os.Rename(jpeg, strings.TrimSuffix(imgpath, ext)+".jpg")
	return
}
