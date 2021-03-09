package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	fatomic "github.com/natefinch/atomic"
)

//Registry holds the registry
var Registry map[string]string

func init() {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		log.Fatal("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := os.ReadFile(registryFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot read file '%s' (BROCADE_REGISTRY environment variable)\n", registryFile), err)
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		log.Fatal(fmt.Sprintf("registry file '%s' does not contain valid JSON.\nUse http://jsonlint.com/\n", registryFile), err)
	}

}

//SetRegistry set a value to a key in the registry
func SetRegistry(key, value string) error {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		return fmt.Errorf("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := os.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("cannot read file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		return fmt.Errorf("registry file `%s` does not contain valid JSON.\nUse http://jsonlint.com/", registryFile)
	}
	Registry[key] = value
	r, err := json.Marshal(Registry)
	if err != nil {
		return fmt.Errorf("cannot marshal to valid JSON: %s", err.Error())
	}
	err = fatomic.WriteFile(registryFile, bytes.NewReader(r))
	if err != nil {
		return fmt.Errorf("cannot write file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	return nil
}
