package document

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	perror "brocade.be/pbladng/lib/error"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
)

func NewImage(s string, lineno int, cpright string, doc *pstructure.Document, dir string, alt string, alts map[string]*pstructure.Image) (err error) {
	name, legend, ok := strings.Cut(s, ".jpg")
	if !ok {
		err = perror.Error("image-jpg", lineno, "line should contain .jpg")
		return
	}
	if strings.TrimSpace(alt) != "" {
		err = perror.Error("image-alt", lineno, "image alt should be empty")
		return
	}

	if name == "" {
		err = perror.Error("image-name", lineno, "name is empty")
		return
	}
	name += ".jpg"
	fname, e := FindImage(name, doc, dir)
	if e != nil {
		err = perror.Error("image-unknown", lineno, e.Error())
		return
	}

	for _, img := range alts {
		if fname == img.Fname {
			err = perror.Error("image-double", lineno, "image also used on line `"+strconv.Itoa(img.Lineno)+"`")
			return
		}
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
		if cpright == "" {
			cpright = pregistry.Registry["copyright-default"].(string)
		}
		copyright = cpright
	}

	if copyright == "" {
		err = perror.Error("image-copyright", lineno, "no copyright for `"+name+"`")
		return
	}

	if len(alts) > 25 {
		err = perror.Error("image-limit", lineno, "too many images")
		return
	}
	image := pstructure.Image{
		Name:      string(rune(97 + len(alts))),
		Legend:    legend,
		Copyright: copyright,
		Fname:     fname,
		Lineno:    lineno,
	}
	c := doc.LastChapter()
	if c == nil {
		return perror.Error("illegal-image1", lineno, "image before chapter")
	}
	t := c.LastTopic()
	if t == nil {
		return perror.Error("illegal-image2", lineno, "image before topic")
	}
	if t.Type == "mass" {
		return perror.Error("illegal-image3", lineno, "images are not allowed in type `mass`")
	}
	t.Images = append(t.Images, &image)
	alts[image.Name] = &image
	return
}

func ImageStore(sdir string, tdir string) (err error) {
	// images with extension .jpg or .jpeg
	// unique identifier: [a-zA-Z0-9]+ (lowercase)
	if tdir == "" {
		tdir = pfs.FName("workspace")
	}
	if sdir == "" {
		sdir = tdir
	}
	type img struct {
		dir  string
		info fs.DirEntry
		name string
		id   string
	}

	jpegs := make([]img, 0)

	fn := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}
		k := strings.IndexAny(name, "-_.")
		index := strings.ToLower(name[:k])
		pname := index + "-" + name[k+1:]
		pname = strings.TrimSuffix(pname, filepath.Ext(pname)) + ".jpg"
		jpegs = append(jpegs, img{
			dir:  filepath.Dir(path),
			info: info,
			name: pname,
			id:   index,
		})
		return nil
	}
	err = filepath.WalkDir(sdir, fn)
	if err != nil {
		err = fmt.Errorf("error walking `%s`: %s", sdir, err)
		return
	}
	match := make(map[string]bool)

	for _, jpg := range jpegs {
		path := filepath.Join(jpg.dir, jpg.info.Name())
		if jpg.id == "" {
			err = fmt.Errorf("invalid id for image `%s`", path)
			return
		}
		if match[jpg.id] {
			err = fmt.Errorf("double id `%s` for image `%s`", jpg.id, path)
			return
		}
		_, e := os.ReadFile(path)
		if e != nil {
			err = fmt.Errorf("cannot read image `%s`: `%s`", path, e.Error())
			return
		}

		sfile := path
		tfile := filepath.Join(tdir, jpg.name)
		if bfs.SameFile(sfile, tfile) {
			continue
		}
		if filepath.Dir(sfile) == filepath.Dir(tfile) {
			err = os.Rename(sfile, tfile)
			if err != nil {
				err = fmt.Errorf("cannot rename `%s` to `%s`", sfile, tfile)
				return
			}
			continue
		}

		err = bfs.CopyFile(sfile, tfile, "", false)
		if err != nil {
			err = fmt.Errorf("cannot copy `%s` to `%s`", sfile, tfile)
			return
		}
	}
	return err
}

func FindImage(index string, doc *pstructure.Document, dir string) (fname string, err error) {

	bindex := index
	index = strings.TrimSuffix(index, ".jpg")

	if index == "" {
		return "", fmt.Errorf("reference to image should not be empty")
	}

	indexen := strings.SplitN(index, "-", -1)

	if strings.TrimLeft(index, "abcdefghijklmnopqrstuvwxyz1234567890") != "" {
		return "", fmt.Errorf("invalid reference to image `%s`", index)
	}

	dirs := []string{dir, pfs.FName("workspace"), pfs.FName(fmt.Sprintf("archive/%d/%02d", doc.Year, doc.Week))}
	index = ""
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		index := ""
		for _, inx := range indexen {
			index = index + "-" + inx
			index = strings.TrimPrefix(index, "-")
			globber := filepath.Join(dir, index+"[_.\\-]*jpg")
			matches, e := filepath.Glob(globber)
			if e != nil {
				return "", fmt.Errorf("cannot glob `%s`: %s", globber, e.Error())
			}
			if len(matches) > 1 {
				return "", fmt.Errorf("too many image results `%s` for `%s`", strings.Join(matches, ", "), index)
			}
			if len(matches) == 1 {
				bfs.CopyFile(matches[0], pfs.FName("workspace"), "", false)
				return matches[0], nil
			}
		}
	}
	return "", fmt.Errorf("no image found for `%s`", bindex)

}
