package registry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"time"

	fatomic "github.com/natefinch/atomic"
)

// Registry holds the registry
var Registry map[string]any

func init() {
	registry := make(map[string]any)
	Registry = make(map[string]any)
	registryFile := os.Getenv("MY_REGISTRY")
	if registryFile == "" {
		Registry["error"] = "MY_REGISTRY environment variable is not defined"
		return
	}
	info, err := os.Stat(registryFile)
	if err == nil && info.IsDir() {
		Registry["error"] = fmt.Sprintf("MY_REGISTRY `%s` points to a directory. It should be a file.", registryFile)
		return
	}
	b := make([]byte, 0)
	if !errors.Is(err, fs.ErrNotExist) {
		b, err = os.ReadFile(registryFile)
		if err != nil {
			Registry["error"] = fmt.Sprintf("cannot read file '%s' (MY_REGISTRY environment variable)", registryFile)
			return
		}
	}
	if len(b) == 0 {
		b = []byte("{}")
		err = fatomic.WriteFile(registryFile, bytes.NewReader(b))
		if err != nil {
			Registry["error"] = fmt.Sprintf("cannot initialise file '%s' (MY_REGISTRY environment variable)", registryFile)
			return
		}
	}
	err = json.Unmarshal(b, &registry)
	if err != nil {
		Registry["error"] = fmt.Sprintf("cannot unmarshal JSON file '%s': '%s')", registryFile, err.Error())
		return
	}
	delete(registry, "error")
	rp, ok := registry["pblad"]
	if ok {
		Registry = rp.(map[string]any)
		_, ok := Registry["valid-until"]
		if !ok {
			Registry["error"] = fmt.Sprintf("No `valid-until` key in registry [%s]", registryFile)
			return
		}
		now := time.Now()
		validuntil := Registry["valid-until"].(string)
		if now.Format(time.RFC3339) > validuntil {
			Registry["error"] = fmt.Sprintf("`valid-until` key says: registry is not up to date [%s]", registryFile)
			return
		}
	} else {
		Registry["error"] = fmt.Sprintf("no \"pblad\" subscript in JSON file '%s')", registryFile)
		return
	}
}
