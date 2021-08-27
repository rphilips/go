package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var fsNoutf8Cmd = &cobra.Command{
	Use:     "noutf8",
	Short:   "Search for non-UTF8 sequences",
	Long:    `This command searches for non-UTF8 byte sequences`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng fs noutf8 *.m cwd=../catalografie`,
	RunE:    fsNoutf8,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsNoutf8Cmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsNoutf8Cmd.Flags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
	fsCmd.AddCommand(fsNoutf8Cmd)
}

func fsNoutf8(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	ask := false

	if len(args) == 0 {
		ask = true
		for {
			fmt.Print("File/directory          : ")
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

	fnoutf8 := func(reader io.Reader) (result int, err error) {
		var breader *bufio.Reader
		if reader == nil {
			breader = bufio.NewReader(os.Stdin)
		} else {
			breader = bufio.NewReader(reader)
		}
		data, err := ioutil.ReadAll(breader)
		if err != nil {
			return
		}
		if bytes.ContainsRune(data, 0) {
			return 0, nil
		}
		if utf8.Valid(data) {
			return -1, nil
		}
		if reader != nil {
			return 1, nil
		}
		for i, line := range bytes.Split(data, []byte{10}) {
			if utf8.Valid(line) {
				continue
			}
			return i, nil
		}
		return -1, nil
	}

	if len(args) == 1 && args[0] == "-" {
		r, err := fnoutf8(nil)
		result := "Valid UTF-8"
		if r > -1 {
			result = fmt.Sprintf("Not valid UTF-8 (line: %d)", 1+r)
		}
		Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	files := make([]string, 0)
	files, err := glob(Fcwd, args, Frecurse, Fpattern, true, false, false)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
			return nil
		}
		msg := make(map[string][]string)
		msg["noutf8"] = files
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	fn := func(n int) (interface{}, error) {
		src := files[n]

		in, err := os.Open(src)
		if err != nil {
			return nil, err
		}
		defer in.Close()
		result, err := fnoutf8(in)
		return result, err
	}
	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	type rtt struct {
		All []string            `json:"all"`
		Ext map[string][]string `json:"ext"`
	}
	rt := rtt{}
	rt.All = make([]string, 0)
	rt.Ext = make(map[string][]string)
	for i, r := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.noutf8"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		if !r.(bool) {
			continue
		}
		ext := path.Ext(src)
		rt.All = append(rt.All, src)
		lst := rt.Ext[ext]
		if len(lst) == 0 {
			lst = make([]string, 0)
		}
		lst = append(lst, src)
		rt.Ext[ext] = lst
	}
	if len(errs) == 0 {
		errs = nil
	}
	lst := rt.All
	sort.Strings(lst)
	rt.All = lst

	for ext, lst := range rt.Ext {
		sort.Strings(lst)
		rt.Ext[ext] = lst
	}

	Fmsg = qreport.Report(rt, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
