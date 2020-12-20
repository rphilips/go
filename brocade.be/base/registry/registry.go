package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	fatomic "github.com/natefinch/atomic"
)

var Registry map[string]string

func init() {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		log.Fatal("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := ioutil.ReadFile(registryFile)
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot read file '%s' (BROCADE_REGISTRY environment variable)\n", registryFile), err)
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		log.Fatal(fmt.Sprintf("registry file '%s' does not contain valid JSON.\nUse http://jsonlint.com/\n", registryFile), err)
	}

}

func SetRegistry(key, value string) error {
	registryFile := os.Getenv("BROCADE_REGISTRY")
	if registryFile == "" {
		return fmt.Errorf("BROCADE_REGISTRY environment variable is not defined")
	}
	b, err := ioutil.ReadFile(registryFile)
	if err != nil {
		return fmt.Errorf("Cannot read file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	err = json.Unmarshal(b, &Registry)
	if err != nil {
		return fmt.Errorf("Registry file `%s` does not contain valid JSON.\nUse http://jsonlint.com/\n", registryFile)
	}
	Registry[key] = value
	r, err := json.Marshal(Registry)
	if err != nil {
		fmt.Errorf("Cannot marshal to valid JSON: %s", err.Error())
	}
	err = fatomic.WriteFile(registryFile, bytes.NewReader(r))
	if err != nil {
		fmt.Errorf("Cannot write file `%s` (BROCADE_REGISTRY environment variable): %s", registryFile, err.Error())
	}
	return nil
}
