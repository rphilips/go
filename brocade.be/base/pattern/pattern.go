package pattern

import (
	"runtime"
	"strconv"
	"sync"

	qregistry "brocade.be/base/registry"
)

// Number beheert een aantal tokens: enkel het aantal is belangrijk
// De functie heeft 3 functies terug:
//    `borrow` laat toe om een token op te vragen
//    `release` brengt een token terug
//    `finish` ruimt op
//    Elke `borrow` moet uiteindelijk worden gevolgd door een `release`
func Number(nrTokens int) (borrow func(), release func(), finish func()) {
	maxopen, _ := qregistry.Registry["qtechng-max-parallel"]
	num := runtime.GOMAXPROCS(-1)
	if maxopen != "" {
		n, e := strconv.Atoi(maxopen)
		if e == nil {
			num = n
		}
	}
	if nrTokens > num || nrTokens < 1 {
		nrTokens = num
	}
	tokens := make(chan int, nrTokens)

	for i := 1; i <= nrTokens; i++ {
		tokens <- 0
	}

	borrow = func() {
		<-tokens
	}

	release = func() {
		tokens <- 0
	}

	finish = func() {
		close(tokens)
	}
	return
}

// Task stands for a task to be executed by AsyncChan
type Task interface {
	ID() string                    // return an identification
	Payload() interface{}          // return data to be handled
	Mission() (interface{}, error) // is the mission to be executed
}

// Result stands for a result per Task
type Result struct {
	ID    string      // identification of the task
	Extra interface{} // extra result information
	Err   error       // en eventual error
}

// AsyncChan handles a stream of task asynchronously and returns the result in another stream
func AsyncChan(input <-chan Task, maxopen int) <-chan Result {
	var wg sync.WaitGroup
	out := make(chan Result)
	borrow, release, finish := Number(maxopen)
	for task := range input {
		wg.Add(1)
		go func(task Task) {
			defer wg.Done()
			borrow()
			defer release()
			id := task.ID()
			res, err := task.Mission()
			out <- Result{id, res, err}
		}(task)
	}
	go func() {
		wg.Wait()
		close(out)
		finish()
	}()
	return out
}
