package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"brocade.be/base/parallel"
	"brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var testStressCmd = &cobra.Command{
	Use:   "stress",
	Short: "Stress test",
	Long: `Perform a IIIF stress test.
	The argument is the number of request to perform.
	With the --rais flag the RAIS image server is tested,
	without this flag (default) the default Apache webserver is tested`,
	Args:    cobra.ExactArgs(1),
	Example: "iiiftool test stress 100",
	RunE:    testStress,
}

var Frais bool

const testId = "e0f4d5d32a3dd5a341ec84a2ae8e9c69e2666fca"

func init() {
	testCmd.AddCommand(testStressCmd)
	testStressCmd.PersistentFlags().BoolVar(&Frais, "rais", false, "Test RAIS IIIF image server instead of Apache")
}

func testStress(cmd *cobra.Command, args []string) error {

	URL := registry.Registry["web-base-url"] + registry.Registry["iiif-base-url"]

	if !Frais {
		URL = URL + "/index.phtml?id=" + testId + "&file=00000001.jp2"
	} else {
		URL = URL + "/" + testId + "00000001.jp2/full/max/0/default.jpg"
	}

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
