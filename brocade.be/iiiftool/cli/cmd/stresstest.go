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

var stresstestCmd = &cobra.Command{
	Use:   "stresstest",
	Short: "stresstest",
	Long: `Perform a IIIF stresstest.
	The argument is the number of request to perform`,
	Args:    cobra.ExactArgs(1),
	RunE:    stresstest,
	Example: "iiiftool stresstest 100",
}

func init() {
	rootCmd.AddCommand(stresstestCmd)
}

func stresstest(cmd *cobra.Command, args []string) error {

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
	fmt.Println(len(result), "requests handled in", diff)

	return nil
}
