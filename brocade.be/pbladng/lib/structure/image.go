package structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	perror "brocade.be/pbladng/lib/error"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

type Image struct {
	Name      string
	Legend    string
	Copyright string
	Fname     string
	Letter    string
	Lineno    int
}

func NewImage(s string, copyright string, lineno int, dirs []string) (image *Image, err error) {
	found := rjpg.FindStringIndex(s)
	if found == nil {
		err = perror.Error("image-jpg", lineno, "line should contain .jpg")
		return
	}
	name := strings.TrimSpace(s[:found[0]])
	legend := ""
	if found[1] < len(s) {
		legend = strings.TrimSpace(s[found[1]:])
	}
	if name == "" {
		err = perror.Error("image-name", lineno, "name is empty")
		return
	}

	fname, e := FindImage(name+".jpg", dirs)
	if e != nil {
		err = perror.Error("image-unknown", lineno, e.Error())
		return
	}

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
		copyright = pregistry.Registry["copyright-default"].(string)
	}

	if copyright == "" {
		err = perror.Error("image-copyright", lineno, "no copyright for `"+name+"`")
		return
	}

	image = &Image{
		Name:      name,
		Legend:    legend,
		Copyright: copyright,
		Fname:     fname,
		Lineno:    lineno,
	}
	return
}

func FindImage(index string, dirs []string) (fname string, err error) {

	bindex := index
	index = strings.ToLower(strings.TrimSuffix(index, ".jpg"))

	if index == "" {
		return "", fmt.Errorf("reference to image should not be empty")
	}

	indexen := strings.SplitN(index, "-", -1)

	index = ""
	fdir := ""
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		if fdir == "" {
			fdir = dir
		}
		globber := filepath.Join(dir, "*.jpg")
		matches, e := filepath.Glob(globber)
		if e != nil {
			return "", fmt.Errorf("cannot glob `%s`: %s", globber, e.Error())
		}
		if len(matches) == 0 {
			continue
		}
		for _, fn := range matches {
			base := filepath.Base(fn)
			base = strings.TrimSuffix(base, ".jpg")
			lbase := "-" + strings.ToLower(base) + "-"
			fname = fn
			ok := false
			for _, inx := range indexen {
				ok = strings.Contains(lbase, "-"+inx+"-")
				if !ok {
					break
				}
			}
			if ok {
				if dir != fdir {
					bfs.CopyFile(fname, fdir, "", false)
				}
				return
			}
		}
	}
	return "", fmt.Errorf("no image found for `%s`", bindex)
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
	worker := pregistry.Registry["image-resize-exe"].([]any)
	keys := map[string]string{"source": imgpath, "kbsize": strconv.Itoa(kbsize), "target": small}
	out, err := ptools.Launch(worker, keys, "", true, false)
	if err != nil {
		err = fmt.Errorf("%s:\n%s", err.Error(), string(out))
		return
	}
	_, err = os.Stat(small)
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
	worker := pregistry.Registry["image-resize-exe"].([]any)
	keys := map[string]string{"source": imgpath, "kbsize": pregistry.Registry["image-size-kb"].(string), "target": jpeg}
	out, err := ptools.Launch(worker, keys, "", true, false)
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
