package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsRegistryCmd = &cobra.Command{
	Use:     "registry",
	Short:   "Searches for the use of brocade registry values",
	Long:    `Searches for the use of brocade registry values in go and python files`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs registry cwd=../catalografie`,
	RunE:    fsRegistry,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsRegistryCmd)
}

func fsRegistry(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)
	ask := false
	if len(args) == 0 {
		ask = true
		for {
			fmt.Print("File/directory        : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			args = append(args, text)
		}
		if len(args) == 0 {
			return nil
		}
	}

	if ask && !Frecurse {
		fmt.Print("Recurse ?               : <n>")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if text == "" {
			text = "n"
		}
		if strings.ContainsAny(text, "jJyY1tT") {
			Frecurse = true
		}
	}

	if ask && len(Fpattern) == 0 {
		for {
			fmt.Print("Pattern on basename     : ")
			text, _ := reader.ReadString('\n')
			text = strings.TrimSuffix(text, "\n")
			if text == "" {
				break
			}
			Fpattern = append(Fpattern, text)
			if text == "*" {
				break
			}
		}
	}

	re1 := regexp.MustCompile("r4_[a-z][A-Za-z0-9_]*")
	re2 := regexp.MustCompile(`\bregistry\.[a-z][A-Za-z0-9_]*`)
	re3 := regexp.MustCompile(`\bregistry[([]\s*['"` + "`" + `]+[a-z][A-Za-z0-9_-]*`)

	if len(Fpattern) == 0 {
		Fpattern = []string{"*"}
	}

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
	if len(args) == 1 && args[0] == "-" {
		r, err := fnreg(nil)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
		if r == nil {
			Fmsg = qreport.Report("No registry value found", err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
		msg := make(map[string][]string)
		regs := make([]string, 0)
		for reg := range r {
			regs = append(regs, reg)
		}
		sort.Strings(regs)
		msg["registry"] = regs
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	files := make([]string, 0)
	files, err := glob(Fcwd, []string{"."}, true, Fpattern, true, false, true)

	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
		msg := make(map[string][]string)
		msg["registry"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
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
