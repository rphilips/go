package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"brocade.be/base/parallel"
	"github.com/spf13/cobra"
)

var testStressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress test",
	Long: `Perform a IIIF stress test.
	The argument is the number of request to perform`,
	Args:    cobra.ExactArgs(1),
	Example: "iiiftool test stress 100",
	RunE:    testStress,
}

func init() {
	testCmd.AddCommand(testStressCmd)
}

func testStress(cmd *cobra.Command, args []string) error {

	URL := "https://dev.anet.be/iiif/index.phtml?id=tg:uakagaab:7188&file=manifest.json"

	requests, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("iiiftool ERROR: invalid number of requests: %s", err)
		return nil
	}

	fn := func(n int) (interface{}, error) {
		response, err := http.Get(URL)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()
		return response.Body, nil
	}

	start := time.Now()
	result, errors := parallel.NMap(requests, -1, fn)
	end := time.Now()
	for _, e := range errors {
		if e != nil {
			log.Fatalf("iiiftool ERROR: error opening URL: %v", e)
		}
	}

	diff := end.Sub(start)
	fmt.Println(len(result), "requests to", URL, "\nhandled in", diff)

	return nil
}
