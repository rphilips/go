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
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	qfs "brocade.be/base/fs"
	"brocade.be/goyo/cli/cmd"
	qyottadb "brocade.be/goyo/lib/yottadb"
	ydb "lang.yottadb.com/go/yottadb"
)

var buildTime string
var goVersion string
var buildHost string

var RoDir = ""

func main() {
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

	defer qyottadb.Exit()
	RoDir = Setup()
	defer os.RemoveAll(RoDir)

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "repl")
	}
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
					if !ok && line != "goyo" {
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
					if len(argums) == 0 || argums[0] != "goyo" {
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

	rand.Seed(time.Now().UTC().UnixNano())
	cmd.Execute(buildTime, goVersion, buildHost, os.Args)
	exec.Command("stty", "sane").Run()

}

func Setup() string {
	dir, err := ioutil.TempDir("", "goyo.")
	if err != nil {
		log.Fatal(err)
	}

	ydbaccess := `ydbaccess ; entry points to access YottaDB
 quit
 ;
get(var,value)
 set value=@var
 quit
 ;
kill(var)
 kill @var
 quit
 ;
lock(var)
 lock @var
 quit
 ;
order(var,value)
 set value=$order(@var)
 quit
 ;
query(var,value)
 set value=$query(@var)
 quit
 ;
set(var,value)
 set @var=value
 quit
 ;
xecutew(var,err)
 new %endofexec,$etrap,$estack
 set %iopipe="/tmp/goyo.pipe"
 o %iopipe:fifo
 u %iopipe
 s %endofexec="<[<end>]>"
 set $etrap="w %endofexec,! c %iopipe quit:$estack  set $ecode="""",err=$zstatus quit:$quit +err quit"
 xecute var
 w %endofexec,!
 c %iopipe
 quit

xecute(var,err)
 new $etrap,$estack
 set $etrap="quit:$estack  set $ecode="""",err=$zstatus quit:$quit +err quit"
 xecute var
 quit
 ;
`

	mfile := filepath.Join(dir, "ydbaccess.m")
	qfs.Store(mfile, ydbaccess, "process")

	ydb_routines := os.Getenv("ydb_routines")
	zro := ydb_routines
	if zro != "" {
		zro = " "
	}
	zro += dir
	err = ydb.SetValE(ydb.NOTTP, nil, zro, "$ZRO", nil)
	if err != nil {
		log.Fatal(err)
	}

	ydbaccess_ci := `get     : void get^ydbaccess(I:ydb_char_t*, O:ydb_string_t*)
kill    : void kill^ydbaccess(I:ydb_char_t*)
lock    : void lock^ydbaccess(I:ydb_char_t*)
order   : void order^ydbaccess(I:ydb_char_t*, O:ydb_string_t*)
query   : void query^ydbaccess(I:ydb_char_t*, O:ydb_string_t*)
set     : void set^ydbaccess(I:ydb_char_t*, I:ydb_string_t*)
xecute  : void xecute^ydbaccess(I:ydb_string_t*, O:ydb_string_t*)
xecutew : void xecutew^ydbaccess(I:ydb_string_t*, O:ydb_string_t*)
`
	mci := filepath.Join(dir, "ydbaccess.ci")
	qfs.Store(mci, ydbaccess_ci, "process")
	os.Setenv("GOYO_DIR", dir)
	return dir
}
