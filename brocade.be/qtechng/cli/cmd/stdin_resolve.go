package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qsource "brocade.be/qtechng/lib/source"
	"github.com/spf13/cobra"
)

var stdinResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "replace i4/r4/l4/m4 constructions",
	Long:  `replace i4/r4/l4/m4 constructions`,
	Example: `
  qtechng stdin resolve --csv=1,3,4`,
	Args: cobra.NoArgs,
	RunE: stdinResolve,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "B",
	},
}

var Fcsv string

func init() {
	stdinResolveCmd.Flags().StringVar(&Fcsv, "csv", "", "source column, target column")
	stdinResolveCmd.PersistentFlags().StringVar(&Frilm, "rilm", "", "specify the substitutions")
	stdinCmd.AddCommand(stdinResolveCmd)
}

func stdinResolve(cmd *cobra.Command, args []string) (err error) {
	qpath := -1
	csource := -1
	ctarget := -1
	if Frilm == "" {
		Frilm = "rilm"
	}

	if Fcsv != "" {
		parts := strings.SplitN(Fcsv, ",", -1)
		if len(parts) == 2 {
			parts = append(parts, parts[1])
		}
		switch len(parts) {
		case 0, 1, 2:
			err := &qerror.QError{
				Ref:  []string{"stdin.resolve.invalidcsv1"},
				Type: "Error",
				Msg:  []string{"csv flag should be of the form `--csv=i,j` or `--csv=i,j,k`"},
			}
			return err
		case 3:
			x, err := strconv.Atoi(strings.TrimSpace(parts[0]))
			if err != nil {
				return err
			}
			if x < 1 {
				err := &qerror.QError{
					Ref:  []string{"stdin.resolve.invalidcsv2"},
					Type: "Error",
					Msg:  []string{"Numbers in `--csv` flag should be greater dan 0"},
				}
				return err
			}
			qpath = x
			x, err = strconv.Atoi(strings.TrimSpace(parts[1]))
			if err != nil {
				return err
			}
			if x < 1 {
				err := &qerror.QError{
					Ref:  []string{"stdin.resolve.invalidcsv3"},
					Type: "Error",
					Msg:  []string{"Numbers in `--csv` flag should be greater dan 0"},
				}
				return err
			}
			csource = x
			x, err = strconv.Atoi(strings.TrimSpace(parts[2]))
			if err != nil {
				return err
			}
			if x < 1 {
				err := &qerror.QError{
					Ref:  []string{"stdin.resolve.invalidcsv4"},
					Type: "Error",
					Msg:  []string{"Numbers in `--csv` flag should be greater dan 0"},
				}
				return err
			}
			ctarget = x
		default:
			err := &qerror.QError{
				Ref:  []string{"stdin.resolve.invalidcsv5"},
				Type: "Error",
				Msg:  []string{"csv flag should be of the form `--csv=i,j` or `--csv=i,j,k`"},
			}
			return err
		}
	}

	reader := bufio.NewReader(os.Stdin)

	output := os.Stdout
	if Fstdout != "" {
		var err error
		output, err = os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer output.Close()
	}

	for {
		eol := ""
		delim := ""
		a, err := reader.ReadString('\n')
		if strings.HasSuffix(a, "\n") {
			a = strings.TrimSuffix(a, "\n")
			eol = "\n"
		}
		if strings.HasSuffix(a, "\r") {
			a = strings.TrimSuffix(a, "\r")
			eol = "\r" + eol
		}
		if a == "" {
			output.WriteString(a)
			output.WriteString(eol)
			if err == io.EOF {
				break
			}
			continue
		}
		if delim == "" && qpath > 0 {
			for _, r := range a {
				if r == 124 || r > 128 {
					delim = string(r)
					break
				}
			}
		}
		switch delim {
		case "":
			r, e := resolve(a, "")
			output.WriteString(r)
			output.WriteString(eol)
			if e != nil {
				os.Stderr.WriteString("ERROR:\n")
				os.Stderr.WriteString(a)
				os.Stderr.WriteString("\n===\n")
				os.Stderr.WriteString(e.Error())
				os.Stderr.WriteString("\n\n")
			}
			if err == io.EOF {
				break
			}
		default:
			parts := strings.SplitN(a, delim, -1)
			for len(parts) < csource {
				parts = append(parts, "")
			}
			for len(parts) < ctarget {
				parts = append(parts, "")
			}
			qp := parts[qpath-1]
			source := parts[csource-1]
			r, e := resolve(source, qp)
			parts[ctarget-1] = r
			output.WriteString(strings.Join(parts, delim))
			output.WriteString(eol)
			if e != nil {
				os.Stderr.WriteString("ERROR:\n")
				os.Stderr.WriteString(a)
				os.Stderr.WriteString("\n===\n")
				os.Stderr.WriteString(e.Error())
				os.Stderr.WriteString("\n\n")
			}
			if err == io.EOF {
				break
			}
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func resolve(s string, qpath string) (result string, err error) {
	fmt.Println(">>>>", s)
	if strings.Contains(qpath, "#") {
		qpath = strings.TrimSpace(strings.SplitN(qpath, "#", 2)[0])
	}
	if qpath == "" {
		qpath = "/qtechng/mumps/qtm4.m"
	}
	r := "0.00"
	source, err := qsource.Source{}.New(r, qpath, true)
	body := []byte(s)

	if !bytes.Contains(body, []byte("4_")) {
		return s, nil
	}
	nature := source.Natures()
	if !nature["text"] {
		return s, nil
	}
	if nature["objfile"] {
		return s, nil
	}
	env := source.Env()
	notreplace := source.NotReplace()
	bufmac := new(bytes.Buffer)
	objectmap := make(map[string]qobject.Object)
	_, err = qsource.ResolveText(env, body, Frilm, notreplace, objectmap, nil, bufmac, "")
	result = string(bufmac.Bytes())
	return
}
