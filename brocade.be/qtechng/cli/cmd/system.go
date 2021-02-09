package cmd

import (
	"strings"

	qregistry "brocade.be/base/registry"
	"github.com/spf13/cobra"
)

var systemCmd = &cobra.Command{
	Use:     "system",
	Short:   "System information",
	Long:    `Command allows for  configuration/testing of setup`,
	Args:    cobra.NoArgs,
	Example: "qtechng system",
}

func init() {
	systemCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	rootCmd.AddCommand(systemCmd)
}

func regMap() map[string]string {
	regmap := map[string]string{
		"brocade-release":   "BP",
		"system-name":       "BP",
		"system-group":      "BP",
		"system-roles":      "BP",
		"private-instname":  "BP",
		"private-role":      "BP",
		"fs-owner-qtechng":  "BWP",
		"lock-dir":          "BP",
		"scratch-dir":       "WBP",
		"m-os-type":         "BP",
		"gtm-rou-dir":       "BP",
		"m-import-auto-exe": "BP",
		"os":                "BP",
		"m-clib":            "BP",
		"web-base-url":      "BP",
	}
	for key, value := range regmap {
		if value != "" && strings.TrimRight(QtechType, value) == QtechType {
			delete(regmap, key)
			continue
		}
		regmap[key] = qregistry.Registry[key]
	}
	for _, item := range regmapqtechng {
		qt := item.qtechtype
		key := item.name
		if qt != "" && strings.TrimRight(QtechType, qt) == QtechType {
			continue
		}
		regmap[key] = qregistry.Registry[key]
	}
	return regmap
}
