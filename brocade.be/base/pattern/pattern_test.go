package pattern

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNumber(t *testing.T) {
	task, done, finish := Number(100)

	var wg sync.WaitGroup
	for i := 1; i <= 2000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			task()
			time.Sleep(100 * time.Millisecond)
			done()
			fmt.Println("Done", i)
		}(i)
	}
	wg.Wait()
	finish()
	return
}
