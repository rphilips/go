package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var testRequestsCmd = &cobra.Command{
	Use:     "requests",
	Short:   "Image requests test",
	Long:    `Perform RAIS IIIF image requests test`,
	Args:    cobra.NoArgs,
	Example: "iiiftool test requests",
	RunE:    testRequests,
}

func init() {
	testCmd.AddCommand(testRequestsCmd)
}

func testRequests(cmd *cobra.Command, args []string) error {

	prefix := registry.Registry["web-base-url"] + registry.Registry["iiif-base-url"]

	URLs := []string{
		prefix + testId + "00000001.jp2/full/max/0/default.jpg",            // default
		prefix + testId + "00000001.jp2/100,200,300,400/max/0/default.jpg", // region
		prefix + testId + "00000001.jp2/full/300,/0/default.jpg",           // size
		prefix + testId + "00000001.jp2/full/max/180/default.jpg",          // rotation
		prefix + testId + "00000001.jp2/full/max/0/bitonal.jpg",            // quality
		prefix + testId + "00000001.jp2/full/max/0/default.png",            //format
	}

	for _, URL := range URLs {
		cmd := exec.Command("curl", "-O", URL)
		fmt.Println("curl", "-O", URL)
		_, err := cmd.Output()
		if err != nil {
			log.Fatalf("iiiftool ERROR: error downloading %s: %v", URL, err)
		}
	}

	return nil
}
