package pool

import (
	"fmt"
	"runtime"
)

// Result types a result of a calculation
type Result struct {
	ID    string
	IDN   int
	Value interface{}
	Err   error
}

// PMap works on all keys in a slice in parallel and returns a result (map indexed on key)
// the key act as an identifier of the job
func PMap(keys []string, fn func(key string) (r interface{}, err error)) (result []Result) {
	switch len(keys) {
	case 0:
		result = nil
	case 1:
		k := keys[0]
		r, err := fn(k)
		result = []Result{Result{k, 0, r, err}}
	default:
		result = make([]Result, 0)
		max := runtime.GOMAXPROCS(-1)
		if maxopen := len(keys); maxopen < max {
			max = maxopen
		}
		cwork := make(chan string, max)
		cresult := make(chan Result)
		defer close(cresult)

		go func() {
			defer close(cwork)
			for _, key := range keys {
				cwork <- key
			}
		}()

		go func() {
			for key := range cwork {
				go func(k string) {
					r, err := fn(k)
					cresult <- Result{k, 0, r, err}
				}(key)
			}
		}()

		for res := range cresult {
			result = append(result, res)
			if len(result) == len(keys) {
				break
			}
		}
	}
	return result
}

// PSlice works on all keys in a slice in parallel and returns a result (map indexed on key)
// the key act as an identifier of the job
func PSlice(keys []int, fn func(key int) (r interface{}, err error)) (result []Result) {
	switch len(keys) {
	case 0:
		result = nil
	case 1:
		k := keys[0]
		r, err := fn(k)
		result = []Result{Result{"", k, r, err}}
	default:
		result = make([]Result, 0)
		max := runtime.GOMAXPROCS(-1)
		if maxopen := len(keys); maxopen < max {
			max = maxopen
		}
		cwork := make(chan int, max)
		cresult := make(chan Result)
		defer close(cresult)

		go func() {
			for id := range keys {
				cwork <- id
			}
			close(cwork)
		}()

		go func() {
			for id := range cwork {
				go func(k int) {
					r, err := fn(keys[k])
					cresult <- Result{"", k, r, err}
				}(id)
			}
		}()

		for res := range cresult {
			result = append(result, res)
			if len(result) == len(keys) {
				break
			}
		}
	}
	return result
}

// NSlice works on all keys in a slice in parallel and returns a result (map indexed on key)
// the key act as an identifier of the job
func NSlice(n int, fn func(m int) (r interface{}, err error)) (result []Result) {
	switch n {
	case 0:
		result = nil
	case 1:
		k := 0
		r, err := fn(k)
		result = []Result{Result{"", k, r, err}}
	default:
		result = make([]Result, 0)
		max := runtime.GOMAXPROCS(-1)
		if maxopen := n; n < max {
			max = maxopen
		}
		cwork := make(chan int, max)
		cresult := make(chan Result, max)
		defer close(cresult)

		go func(n int) {
			for m := 0; m < n; m++ {
				cwork <- m
			}
			close(cwork)
		}(n)

		go func() {
			for m := range cwork {
				go func(k int) {
					r, err := fn(k)
					cresult <- Result{"", k, r, err}
				}(m)
			}
		}()

		for res := range cresult {
			result = append(result, res)
			n--
			if n == 0 {
				break
			}
		}
	}
	return result
}

// NSelect selects all number sfo rwhich a condition is true
func NSelect(n int, fn func(key int) (istrue bool)) (result []int) {
	switch n {
	case 0:
		result = nil
	case 1:
		if fn(0) {
			return []int{0}
		}
		return nil
	default:
		result = make([]int, 0)
		max := runtime.GOMAXPROCS(-1)
		if maxopen := n; maxopen < max {
			max = maxopen
		}
		cwork := make(chan int, max)
		cresult := make(chan int, max)
		defer close(cresult)

		go func(n int) {
			for m := 0; m < n; m++ {
				cwork <- m
			}
			close(cwork)
		}(n)

		go func() {
			for m := range cwork {
				go func(k int) {
					if fn(k) {
						cresult <- k
					} else {
						cresult <- -1
					}
				}(m)
			}
		}()
		for res := range cresult {
			if res >= 0 {
				result = append(result, res)
				fmt.Println("res:", res, len(cresult))
			}
			n--
			if n == 0 {
				break
			}
		}
	}
	return result
}

// DirPerFile give a directory and a function working on the absolute filename
func DirPerFile(dirname string, fn func(filename string) string) (result map[string]string) {
	return
}
