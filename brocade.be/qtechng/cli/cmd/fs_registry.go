package cmd

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsRegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Searches for the use of brocade registry values",
	Long: `Searches for the use of brocade registry values files

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content


Some remarks:

    - With the '--ask' flag, you can interactively specify the arguments and flags`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs registry *.m cwd=../workspace
qtechng fs registry --ask`,
	RunE: fsRegistry,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsRegistryCmd)
}

func fsRegistry(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor, Fcwd)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-registry-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-registry-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-registry-nofiles")
		return nil
	}

	re1 := regexp.MustCompile("r4_[a-z][A-Za-z0-9_]*")
	re2 := regexp.MustCompile(`\bregistry\.[a-z][A-Za-z0-9_]*`)
	re3 := regexp.MustCompile(`\bregistry[([]\s*['"` + "`" + `]+[a-z][A-Za-z0-9_-]*`)

	fnreg := func(reader io.Reader) (map[string]bool, error) {
		result := make(map[string]bool)
		var breader *bufio.Reader
		if reader == nil {
			breader = bufio.NewReader(os.Stdin)
		} else {
			breader = bufio.NewReader(reader)
		}
		data, err := ioutil.ReadAll(breader)
		if err != nil {
			return result, err
		}

		sdata := strings.ToLower(string(data))
		if !strings.Contains(sdata, "r4_") && !strings.Contains(sdata, "registry") {
			return result, nil
		}
		s1 := re1.FindAllString(sdata, -1)
		s2 := re2.FindAllString(sdata, -1)
		s3 := re3.FindAllString(sdata, -1)
		s1 = append(s1, s2...)
		s1 = append(s1, s3...)
		for _, r := range s1 {
			r = strings.TrimPrefix(r, "r4_")
			r = strings.ReplaceAll(r, "_", "-")
			r = strings.ReplaceAll(r, "'", "")
			r = strings.ReplaceAll(r, "`", "")
			r = strings.ReplaceAll(r, " ", "")
			r = strings.ReplaceAll(r, "\"", "")
			r = strings.TrimPrefix(r, "registry.")
			r = strings.TrimPrefix(r, "registry")
			r = strings.ReplaceAll(r, "(", "")
			r = strings.ReplaceAll(r, "[", "")
			r = strings.TrimRight(r, "-")
			if r != "" {
				result[r] = true
			}
		}

		return result, nil
	}

	fn := func(n int) (interface{}, error) {
		src := files[n]
		in, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		defer in.Close()
		result, err := fnreg(in)
		return result, err
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	registries := make(map[string][]string)
	for i, result := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.registry"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		for r := range result.(map[string]bool) {
			data := registries[r]
			if len(data) == 0 {
				data = make([]string, 0)
			}
			registries[r] = append(data, src)
		}
	}

	msg := make(map[string]map[string][]string)
	msg["registry"] = registries
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
