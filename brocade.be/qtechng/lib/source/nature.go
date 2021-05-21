package source

import (
	"path/filepath"
	"strings"

	qfnmatch "brocade.be/base/fnmatch"
)

// Natures haalt de naturen op van een bestand
//   de volgende naturen zijn gedefinieerd
//   "config", "install", "check", "release"
//   "text"
//   "binary"
//   "bfile", "dfile", "ifile", "lfile", "mfile", "xfile", "dfile"
//   "objectfile"
func (source Source) Natures() map[string]bool {
	if source.natures != nil {
		return source.natures
	}
	s := source.String()
	project := source.project
	srel := source.Rel()
	natures := make(map[string]bool)
	natures["text"] = true

	if srel == "check.py" || srel == "install.py" || srel == "release.py" {
		natures[srel[:len(srel)-3]] = true
		source.natures = natures
		return source.natures
	}
	config, _ := source.project.LoadConfig()

	if project.IsConfig(s) {
		natures["config"] = true
		source.natures = natures
		return source.natures
	}
	binaries := config.Binary
	if len(binaries) != 0 {
		for _, binary := range binaries {
			if qfnmatch.Match(binary, srel) {
				delete(natures, "text")
				natures["binary"] = true
				source.natures = natures
				return source.natures
			}
		}
	}
	ext := filepath.Ext(s)
	if ext != "." && strings.HasPrefix(ext, ".") {
		ext = ext[1:]
		if strings.Contains("bdilmx", ext) {
			notbrocade := config.NotBrocade
			if len(notbrocade) != 0 {
				for _, nb := range notbrocade {
					if qfnmatch.Match(nb, srel) {
						source.natures = natures
						return source.natures
					}
				}
			}
			natures[ext+"file"] = true
			natures["auto"] = true
			if strings.ContainsAny(ext, "blxm") {
				natures["mumps"] = true
			}
		}
		if strings.Contains("dil", ext) {
			natures["objectfile"] = true
		}
	}
	source.natures = natures
	return source.natures
}
