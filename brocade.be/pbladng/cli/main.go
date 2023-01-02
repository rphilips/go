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
	"fmt"
	"os"

	"brocade.be/pbladng/cli/cmd"

	pregistry "brocade.be/pbladng/lib/registry"
)

var buildTime string
var goVersion string
var buildHost string

func main() {
	v, ok := pregistry.Registry["error"]
	if ok {
		fmt.Fprintf(os.Stderr, "error: %s\n", v.(string))
		os.Exit(1)
	}
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "warn")
	}
	cmd.Execute(buildTime, goVersion, buildHost, os.Args)
}
