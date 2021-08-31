// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"brocade.be/qtechng/cli/cmd"

	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
)

var buildTime string
var goVersion string
var buildHost string

func main() {

	var payload *qclient.Payload

	if len(os.Args) > 2 && os.Args[1] == "arg" {
		data := make([]byte, 0)
		mode := os.Args[2]
		var err error
		switch mode {
		case "stdin":
			data, err = io.ReadAll(os.Stdin)
			if err != nil {
				data = nil
				break
			}
		case "file":
			if len(os.Args) < 4 {
				data = nil
				break
			}
			fname := os.Args[3]
			var err error
			file, err := os.Open(fname)
			if err != nil {
				data = nil
				break
			}
			defer file.Close()
			data, err = io.ReadAll(file)
			if err != nil {
				data = nil
				break
			}
		case "json":
			if len(os.Args) < 4 {
				data = nil
				break
			}
			data = []byte(os.Args[3])

		case "url":
			if len(os.Args) < 4 {
				data = nil
				break
			}
			jarg := os.Args[3]
			resp, err := http.Get(jarg)
			if err != nil {
				break
			}
			defer resp.Body.Close()
			data, err = io.ReadAll(resp.Body)
			if err != nil {
				data = nil
				break
			}

		}

		args := make([]string, 0)
		args = append(args, os.Args[0])
		if data != nil {
			data = bytes.TrimSpace(data)
			sdata := string(data)
			if !strings.HasPrefix(sdata, "[") {
				lines := strings.SplitN(sdata, "\n", -1)
				ok := false
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					if !ok && line != "qtechng" {
						args = nil
						break
					}
					if !ok {
						ok = true
						continue
					}
					args = append(args, line)
				}
			} else {
				argums := make([]string, 0)
				err := json.Unmarshal(data, &argums)
				if err != nil {
					args = nil
				} else {
					if len(argums) == 0 || argums[0] != "qtechng" {
						args = nil
					} else {
						args = append(args, argums[1:]...)
					}
				}
			}

		}
		if args != nil {
			os.Args = args
		}
	}

	if len(os.Args) > 5 && os.Args[1] == "lock" && os.Args[2] == "run" {
		args := os.Args[3:]
		cmd.LockRunner(args)
		os.Exit(0)
	}

	if len(os.Args) == 1 {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			blocktime := qregistry.Registry["qtechng-block-qtechng"]
			if blocktime != "" {
				h := time.Now()
				t := h.Format(time.RFC3339)
				if strings.Compare(blocktime, t) < 0 {
					blocktime = ""
					qregistry.SetRegistry("qtechng-block-qtechng", "")
				}
			}
			if blocktime != "" {
				l := log.New(os.Stderr, "", 0)
				l.Fatal("Blocked for workstations until: `" + blocktime + "`")
			}
			payload = qclient.ReceivePayload(os.Stdin)
			os.Args = append(os.Args[:1], "--transported")
			os.Args = append(os.Args, payload.Args...)
		}
	}
	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute(buildTime, goVersion, buildHost, payload, os.Args)
}
