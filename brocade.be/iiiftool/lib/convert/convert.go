package convert

import (
	"fmt"
	"sync"
)

func convert(file string, wg *sync.WaitGroup, errors *[]error) {
	defer wg.Done()
	// converteer een file en capteer error
	*errors = append(*errors, fmt.Errorf(file+": error"))
}

// Wrapper for file conversion
func Run(files []string) error {
	var errors []error
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go convert(file, &wg, &errors)
	}
	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("error converting file(s) %s", errors)
	}
	return nil
}
