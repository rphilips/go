package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

var testRequestsCmd = &cobra.Command{
	Use:     "requests",
	Short:   "requests",
	Long:    `Perform a IIIF image requests test`,
	Args:    cobra.NoArgs,
	Example: "iiiftool test requests",
	RunE:    testRequests,
}

func init() {
	testCmd.AddCommand(testRequestsCmd)
}

func testRequests(cmd *cobra.Command, args []string) error {

	URLs := []string{
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/full/max/0/default.jpg",            // default
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/100,200,300,400/max/0/default.jpg", // region
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/full/300,/0/default.jpg",           // size
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/full/max/180/default.jpg",          // rotation
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/full/max/0/bitonal.jpg",            // quality
		"https://dev.anet.be/iiif/e1e53b3d6b74c2e7ed0615ec687e68fdb61de24200000001.jp2/full/max/0/default.png",            //format
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
