package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	qmumps "brocade.be/base/mumps"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var mumpsJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Starts a M stream",
	Long: `This command writes a M global on stdout in JSON format

The argument is a M global reference (can also be in simplified form)`,
	Args: cobra.ExactArgs(1),
	Example: `qtechng mumps json '^BCAT("lvd",100)'
qtechng mumps json BCAT/lvd/100
qtechng mumps json BCAT/lvd/100 --stdout=myfile.txt`,
	RunE: mumpsJSON,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BP",
		"fill-version":   "yes",
	},
}

var Fhtmlsafe bool

func init() {
	mumpsCmd.AddCommand(mumpsJSONCmd)
	mumpsJSONCmd.Flags().BoolVar(&Fhtmlsafe, "htmlsafe", false, "HTML safe encoding")
}

func mumpsJSON(cmd *cobra.Command, args []string) error {

	name := qutil.MName(args[0], true)

	if name == "" {
		Fmsg = qreport.Report(nil, errors.New("empty or invalid global reference"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}

	payload := map[string]string{"mglobal": name}
	action := "d %Action^stdjglo(.RApayload)"

	oreader, _, err := qmumps.Reader(action, payload)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	js, err := io.ReadAll(oreader)
	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}

	dst := os.Stdout

	if Fstdout != "" {
		dst, err = os.Create(Fstdout)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
		defer dst.Close()
	}

	enc := json.NewEncoder(dst)
	enc.SetIndent("", "    ")
	enc.SetEscapeHTML(Fhtmlsafe)
	output := string(js)
	if len(Fjq) != 0 || Fyaml {
		output, err = qutil.Transform(js, Fjq, Fyaml)
		if err != nil {
			Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}
	}

	dec := json.NewDecoder(bytes.NewReader([]byte(output)))
	dec.UseNumber()

	var data interface{}

	for dec.More() {
		if err = dec.Decode(&data); err != nil {
			e := fmt.Errorf("problem with %q: %v", name, err)
			Fmsg = qreport.Report(nil, e, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
			return nil
		}

		if err = enc.Encode(&data); err != nil {
			e := fmt.Errorf("cannot write out tidy %q: %v", name, err)
			Fmsg = qreport.Report(nil, e, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		}
	}

	return nil
}
