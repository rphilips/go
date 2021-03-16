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
	"bufio"
	"encoding/json"
	"io"
	"log"
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

	// if len(os.Args) > 4 && os.Args[1] == "lock" && os.Args[2] == "run" && os.Args[4] != "" {
	// 	x, e := json.Marshal(os.Args[4:])
	// 	if e == nil {
	// 		os.Args[4] = string(x)
	// 		os.Args = os.Args[:5]
	// 	}
	// }
	var payload *qclient.Payload
	if len(os.Args) > 2 && os.Args[1] == "arg" {
		ok := false
		mode := os.Args[2]
		args := make([]string, 0)
		switch mode {
		case "file", "stdin":
			if mode == "file" && len(os.Args) == 3 {
				break
			}
			file := os.Stdin
			if mode == "file" {
				if len(os.Args) == 3 {
					break
				}
				fname := os.Args[3]
				var err error
				file, err = os.Open(fname)
				defer file.Close()
				if err != nil {
					break
				}
			}

			reader := bufio.NewReader(file)
			for {
				a, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					break
				}
				if strings.HasSuffix(a, "\n") {
					a = strings.TrimSuffix(a, "\n")
				}
				if strings.HasSuffix(a, "\r") {
					a = strings.TrimSuffix(a, "\r")
				}
				if a != "" {
					args = append(args, a)
				}
				if err != nil {
					ok = true
					break
				}
			}
		case "json":
			if len(os.Args) == 3 {
				break
			}
			jarg := os.Args[3]
			if !strings.HasPrefix(jarg, "[") {
				break
			}
			err := json.Unmarshal([]byte(jarg), &args)
			if err != nil {
				break
			}
			ok = true
		}
		if ok {
			length := len(os.Args)
			k := 0
			for i, a := range args {
				k = i + 1
				if k < length {
					os.Args[i+1] = a
				} else {
					os.Args = append(os.Args, a)
				}
			}
			if k+1 < length {
				os.Args = os.Args[:k+1]
			}
		}
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
	cmd.Execute(buildTime, goVersion, buildHost, payload)
}
