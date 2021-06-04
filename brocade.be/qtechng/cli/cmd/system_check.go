package cmd

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemCheckCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check system configuration",
	Long:    `Check system configuration`,
	Args:    cobra.NoArgs,
	Example: "  qtechng system check",
	RunE:    systemCheck,
}

func init() {
	systemCmd.AddCommand(systemCheckCmd)
}

func systemCheck(cmd *cobra.Command, args []string) error {
	errmsg := make([]string, 0)
	if qregistry.Registry["qtechng-test"] != "test-entry" {
		errmsg = append(errmsg, `Registry("qtechng-test") missing or wrong value`)
	}
	if QtechType == "" {
		errmsg = append(errmsg, `Registry("qtechng-type") missing or wrong value`)
	}
	if strings.Contains(QtechType, "B") {
		if qregistry.Registry["qtechng-exe"] != "qtechng" {
			errmsg = append(errmsg, `Registry("qtechng-exe") missing or wrong value`)
		} else {
			exe := "qtechng"
			exe, err := exec.LookPath(exe)
			if err != nil {
				errmsg = append(errmsg, `Cannot find Registry("qtechng-exe") in PATH`)
			} else {
				if filepath.Dir(exe) != qregistry.Registry["bindir"] {
					errmsg = append(errmsg, `Registry("qtechng-exe") installed in the wrong directory`)
				}
			}

		}

		if qregistry.Registry["qtechng-max-parallel"] == "" {
			errmsg = append(errmsg, `Registry("qtechng-max-parallel") missing or wrong value`)
		}
		if qregistry.Registry["qtechng-user"] == "" {
			errmsg = append(errmsg, `Registry("qtechng-user") missing or wrong value`)
		}
		if qregistry.Registry["qtechng-unique-ext"] == "" {
			errmsg = append(errmsg, `Registry("qtechng-unique-ext") missing or wrong value`)
		}

		copy := qregistry.Registry["qtechng-copy-exe"]
		if copy == "" {
			errmsg = append(errmsg, `Registry("qtechng-copy-exe") missing or wrong value`)
			args := make([]string, 0)
			err := json.Unmarshal([]byte(copy), &args)
			if err != nil {
				errmsg = append(errmsg, `Registry("qtechng-copy-exe") missing or not JSON valid`)
			} else {
				if len(args) != 0 {
					sync := args[0]
					if sync == "" {
						errmsg = append(errmsg, `Registry("qtechng-copy-exe") should start with an executable`)
					} else {
						_, err := exec.LookPath(sync)
						if err != nil {
							errmsg = append(errmsg, `Registry("qtechng-copy-exe") should start with an executable in PATH`)
						}
					}
				}
			}
		}

		if QtechType == "" {
			errmsg = append(errmsg, `Registry("qtechng-type") missing or wrong value`)
		}
		if strings.Contains(QtechType, "B") {
			if qregistry.Registry["qtechng-exe"] != "qtechng" {
				errmsg = append(errmsg, `Registry("qtechng-exe") missing or wrong value`)
			} else {
				exe := "qtechng"
				exe, err := exec.LookPath(exe)
				if err != nil {
					errmsg = append(errmsg, `Cannot find Registry("qtechng-exe") in PATH`)
				} else {
					if filepath.Dir(exe) != qregistry.Registry["bindir"] {
						errmsg = append(errmsg, `Registry("qtechng-exe") installed in the wrong directory`)
					}
				}

			}

			if qregistry.Registry["qtechng-max-parallel"] == "" {
				errmsg = append(errmsg, `Registry("qtechng-max-parallel") missing or wrong value`)
			}
			if qregistry.Registry["qtechng-unique-ext"] == "" {
				errmsg = append(errmsg, `Registry("qtechng-unique-ext") missing or wrong value`)
			}

			tomumps := qregistry.Registry["m-import-auto-exe"]
			if tomumps == "" {
				errmsg = append(errmsg, `Registry("m-import-auto-exe") missing or wrong value`)
				args := make([]string, 0)
				err := json.Unmarshal([]byte(tomumps), &args)
				if err != nil {
					errmsg = append(errmsg, `Registry("m-import-auto-exe") missing or not JSON valid`)
				} else {
					if len(args) != 0 {
						sync := args[0]
						if sync == "" {
							errmsg = append(errmsg, `Registry("m-import-auto-exe") should start with an executable`)
						} else {
							_, err := exec.LookPath(tomumps)
							if err != nil {
								errmsg = append(errmsg, `Registry("m-import-auto-exe") should start with an executable in PATH`)
							}
						}
					}
				}
			}

		}

	}
	if len(errmsg) == 0 {
		errmsg = append(errmsg, `OK`)
	}

	Fmsg = qreport.Report(errmsg, nil, Fjq, Fyaml, Funquote, Fsilent)
	return nil
}
