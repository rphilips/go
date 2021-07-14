package cmd

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strconv"
	"strings"

	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var stdinResolveCmd = &cobra.Command{
	Use:   "resolve",
	Short: "replace i4/r4/l4/m4 constructions",
	Long:  `replace i4/r4/l4/m4 constructions`,
	Example: `
  qtechng stdin resolve --csv=1,3,4
  qtechng stdin resolve "s x=m4_CO" --csv=1,3,4`,
	Args: cobra.MaximumNArgs(1),
	RunE: stdinResolve,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BP",
	},
}

var Fcsv string
var Fdelim string
var Fencoded bool

func init() {
	stdinResolveCmd.Flags().StringVar(&Fcsv, "csv", "", "qpath column,source column,target column,editfile column")
	stdinResolveCmd.Flags().StringVar(&Frilm, "rilm", "", "specify the substitutions")
	stdinResolveCmd.Flags().StringVar(&Fdelim, "delimiter", "", "specify the delimiter. Default is tab")
	stdinResolveCmd.Flags().BoolVar(&Fencoded, "encode", false, "JSON encoded")
	stdinCmd.AddCommand(stdinResolveCmd)
}

func stdinResolve(cmd *cobra.Command, args []string) (err error) {
	qpath := -1
	csource := -1
	ctarget := -1
	cedit := -1
	delim := "\t"
	if Fdelim != "" {
		delim = Fdelim
	}
	if Fcsv == "" {
		delim = ""
	}
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
		case 3, 4:
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
			if len(parts) == 4 {
				x, err = strconv.Atoi(strings.TrimSpace(parts[3]))
				if err != nil {
					return err
				}
				if x < 1 {
					err := &qerror.QError{
						Ref:  []string{"stdin.resolve.invalidcsv5"},
						Type: "Error",
						Msg:  []string{"Numbers in `--csv` flag should be greater dan 0"},
					}
					return err
				}
				cedit = x
			}

		default:
			err := &qerror.QError{
				Ref:  []string{"stdin.resolve.invalidcsv5"},
				Type: "Error",
				Msg:  []string{"csv flag should be of the form `--csv=i,j` or `--csv=i,j,k`"},
			}
			return err
		}
	}
	var reader *bufio.Reader
	if len(args) == 0 {
		reader = bufio.NewReader(os.Stdin)
	} else {
		reader = bufio.NewReader(strings.NewReader(args[0]))
	}

	output := os.Stdout
	if Fstdout != "" {
		var err error
		output, err = os.Create(Fstdout)
		if err != nil {
			return err
		}
		defer output.Close()
	}
	edfiles := make(map[string]string)

	for {
		eol := ""
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
		switch delim {
		case "":
			r, e := resolve(a, "")
			if Fencoded {
				r = qutil.Encode(r)
			}
			output.WriteString(r)
			output.WriteString(eol)
			if e != nil {
				os.Stderr.WriteString("ERROR:\n")
				os.Stderr.WriteString(a)
				os.Stderr.WriteString("\n===\n")
				os.Stderr.WriteString(e.Error())
				os.Stderr.WriteString("\n\n")
			}
		default:
			parts := strings.SplitN(a, delim, -1)
			for len(parts) < csource {
				parts = append(parts, "")
			}
			source := parts[csource-1]
			for len(parts) < ctarget {
				parts = append(parts, "")
			}
			qp := parts[qpath-1]
			if cedit > -1 {
				for len(parts) < cedit {
					parts = append(parts, "")
				}

				s := editfile(source, edfiles)
				parts[cedit-1] = s
			}
			r, e := resolve(source, qp)
			if Fencoded {
				r = qutil.Encode(r)
			}
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
		}
		if err == io.EOF {
			break
		}
	}
	return nil
}

func resolve(s string, qpath string) (result string, err error) {
	if strings.Contains(qpath, "#") {
		qpath = strings.TrimSpace(strings.SplitN(qpath, "#", 2)[0])
	}
	if qpath == "" {
		qpath = "/qtechng/mumps/qtm4.m"
	}
	r := "0.00"
	source, err := qsource.Source{}.New(r, qpath, true)
	if err != nil {
		return "qpath `" + qpath + "` does not exist", nil
	}
	body := []byte(s)

	if !bytes.Contains(body, []byte("4_")) {
		return s, nil
	}
	nature := source.Natures()
	if !nature["text"] {
		return s, nil
	}
	if nature["objectfile"] {
		return s, nil
	}
	env := source.Env()
	notreplace := source.NotReplace()
	bufmac := new(bytes.Buffer)
	objectmap := make(map[string]qobject.Object)
	_, err = qsource.ResolveText(env, body, Frilm, notreplace, objectmap, nil, bufmac, "", qpath)
	result = bufmac.String()
	return
}

func editfile(s string, edfiles map[string]string) (result string) {
	k := strings.Index(s, "(")
	if k != -1 {
		s = s[:k]
	}
	ed, ok := edfiles[s]
	if ok {
		return ed
	}
	objlst := []string{s}
	objmap := qobject.InfoObjectList("0.00", objlst)
	if len(objmap) != 1 {
		return "?"
	}
	if objmap[s] == nil {
		return "?"
	}
	edo := objmap[s].(*qobject.Uber)

	edfiles[s] = edo.EditFile()
	return edfiles[s]
}
