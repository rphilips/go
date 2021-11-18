package convert

import (
	"fmt"
)

func convert(file string, err chan error) {
	err <- nil
}

// Wrapper for file conversion
func Run(files []string) error {
	for _, file := range files {
		convErr := make(chan error)
		go convert(file, convErr)
		fmt.Println(file)
		err := <-convErr
		if err != nil {
			return fmt.Errorf("error converting file %s: %v", file, err)
		}
	}
	return nil
}
