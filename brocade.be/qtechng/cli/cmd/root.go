package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qlog "brocade.be/qtechng/lib/log"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

// BuildTime defined by compilation
var BuildTime = ""

// GoVersion defined by compilation
var GoVersion = ""

// BuildHost defined by compilation
var BuildHost = ""

// QtechType should be "W", "B" or "P"
var QtechType = qregistry.Registry["qtechng-type"]

// OQtechType original QtechType
var OQtechType = QtechType

var errRoot = ""

// Fcwd Current working directory
var Fcwd string // current working directory

//Fenv environment variables
var Fenv []string

// FUID userid
var FUID string

// Fjoiner joins lists
var Fjoiner string

// Fpayload pointer to payload information
var Fpayload *qclient.Payload

// Fcargo pointer to payoff information
var Fcargo *qclient.Cargo = new(qclient.Cargo)

// Fqdir qpath of a directory in version
var Fqdir string

// Frstripln right trim trailing carriage returns
var Frstripln bool

// Fproject qpath of a project in version
var Fproject string

// Fversion version to use
var Fversion string // version to use

//Fappend appends to file
var Fappend bool

// Fremote true if it should be executed remotely ?
var Fremote bool

// Frecurse recurse through the file tree
var Frecurse bool // recurse ?

// Fbackup recurse through the file tree
var Fbackup bool // backup ?

// Ftransported indicates if command is transported
var Ftransported bool

// Fregexp use regular expressions ?
var Fregexp bool // regexp ?

//Ftolower lowercase strings
var Ftolower bool

//Fsmartcase smartcase in search staat off
var Fsmartcase bool

// Fneedle needles to search for
var Fneedle []string

// Fpattern patterns to select on
var Fpattern []string

// Fqpattern patterns to select on
var Fqpattern []string

// Fforce overrules normal behaviour
var Fforce bool // force ?

// Feditor editor used in development
var Feditor string // editor used in development
// Fmsg result as a string
var Fmsg string // result as a string

// Fjq JSONPath
var Fjq []string

// Ferrid
var Ferrid string

// Fyaml YAML
var Fyaml bool

// Funhex decides if the args are to be unhexed (if starting with `.`)
var Funhex bool

// Fcmpversion version to compare with
var Fcmpversion string

// Ffilesinproject indicates if files in the project need to be selected
var Ffilesinproject bool

// Fnature natures of the sources
var Fnature []string

//Fcu creation user
var Fcu []string

//Fmu modification user
var Fmu []string

//Fctbefore created before
var Fctbefore string

//Fctafter created after
var Fctafter string

//Fmtbefore modified before
var Fmtbefore string

//Fmtafter modified after
var Fmtafter string

//Fperline per line ?
var Fperline bool

//Fsilent no output ?
var Fsilent bool

//Fstdout name of stdout
var Fstdout string

//Ffiletype extension of the file
var Ffiletype string

// Frefname reference to the installation
var Frefname string

// Funquote unquotes JSON
var Funquote bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "qtechng",
	Short:         "Brocade software maintenance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Long: `qtechng maintains the Brocade software:

    - Development
    - Installation
    - Deployment
    - Version Control
	- Management`,
	PersistentPreRunE: preRun,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&FUID, "uid", "", "User ID")
	rootCmd.PersistentFlags().StringVar(&Fstdout, "stdout", "", "Filename containing the result")
	rootCmd.PersistentFlags().BoolVar(&Ftransported, "transported", false, "Indicate if command is transported")
	rootCmd.PersistentFlags().BoolVar(&Funquote, "unquote", false, "Remove JSON escapes in a string")
	rootCmd.PersistentFlags().StringVar(&Fjoiner, "joiner", "", "Join lists with unquote")
	rootCmd.PersistentFlags().MarkHidden("transported")
	rootCmd.PersistentFlags().MarkHidden("uid")

	rootCmd.PersistentFlags().StringVar(&Fcwd, "cwd", "", "Working directory")
	rootCmd.PersistentFlags().BoolVar(&Funhex, "unhex", false, "Unhexify the arguments starting with `.`")
	rootCmd.PersistentFlags().StringVar(&Feditor, "editor", "", "Editor name")
	rootCmd.PersistentFlags().StringArrayVar(&Fjq, "jsonpath", []string{}, "JSONpath")
	rootCmd.PersistentFlags().BoolVar(&Fyaml, "yaml", false, "Convert to YAML")
	rootCmd.PersistentFlags().BoolVar(&Fsilent, "quiet", false, "Silent the output")
	rootCmd.PersistentFlags().StringArrayVar(&Fenv, "env", []string{}, "Environment variable KEY=VALUE")

}

func preRun(cmd *cobra.Command, args []string) (err error) {

	if len(Fenv) != 0 {
		for _, env := range Fenv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 0 {
				continue
			}
			env := strings.TrimSpace(parts[0])
			if env == "" {
				continue
			}
			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}
			os.Setenv(env, value)
		}

	}

	if cmd.Annotations["rstrip-trailing-crlf"] == "yes" {
		Frstripln = true
	}
	if Funhex && !Ftransported {
		for count, arg := range args {
			if strings.HasPrefix(arg, ".") {
				arg = arg[1:]
				decoded, e := hex.DecodeString(arg)
				if e == nil {
					args[count] = string(decoded)
				}
			}
		}
	}

	// Cwd
	Fcwd, err = checkCwd(Fcwd)
	if err != nil {
		return
	}
	// remote
	if strings.ContainsRune(QtechType, 'B') {
		Fremote = false
	}

	// version
	if cmd.Annotations["fill-version"] == "yes" {
		fillVersion()
	}

	// qdir
	if cmd.Annotations["fill-qdir"] == "yes" {
		fillQdir()
	}

	errRoot = nameCmd(cmd)

	// User
	if FUID == "" {
		FUID = checkUID(FUID)
	}
	if !Ftransported && FUID != "" {
		ok := true
		for _, x := range os.Args {
			if !strings.HasPrefix(x, "--uid=") {
				continue
			}
			ok = false
			break
		}
		if ok {
			os.Args = append(os.Args, "--uid="+FUID)
		}
	}

	// conditions

	if !Fremote && !strings.ContainsRune(QtechType, 'B') && cmd.Annotations["always-remote"] == "yes" {
		Fremote = true
	}
	if !Fremote && !strings.ContainsRune(QtechType, 'B') && strings.ContainsRune(QtechType, 'W') && cmd.Annotations["always-remote-onW"] == "yes" {
		Fremote = true
	}
	if Fremote && cmd.Annotations["remote-allowed"] != "yes" && cmd.Annotations["always-remote"] != "yes" && (!strings.ContainsRune(QtechType, 'W') || cmd.Annotations["always-remote-onW"] != "yes") {
		err = &qerror.QError{
			Ref: []string{errRoot + "remote"},
			Msg: []string{"Command is not allowed on remote server"},
		}
	}

	allowed, ok := cmd.Annotations["with-qtechtype"]
	if ok {
		ok = false
		for _, char := range QtechType {
			ok = strings.ContainsRune(allowed, char)
			if ok {
				break
			}
		}
		if !ok {
			err = &qerror.QError{
				Ref: []string{errRoot + "QtechType"},
				Msg: []string{fmt.Sprintf("Command not allowed with qtechng-type `%s`", QtechType)},
			}
		}
	}

	withforce := cmd.Annotations["with-force"]
	if withforce == "yes" && !Fforce {
		err = &qerror.QError{
			Ref: []string{errRoot + "force"},
			Msg: []string{"Command should be run with --force=true"},
		}
	}

	// Jq
	if len(Fjq) != 0 && !Ftransported {
		_, err = qutil.Transform(nil, Fjq, false)
		if err != nil {
			err = &qerror.QError{
				Ref: []string{errRoot + "jsonpath"},
				Msg: []string{fmt.Sprintf("JSONpath error: " + err.Error())},
			}
		}
	}

	qlog.Log(BuildTime, FUID, Fversion, os.Args[1:])

	if err != nil {
		return
	}

	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(buildTime string, goVersion string, buildHost string, payload *qclient.Payload, args []string) {
	defer qlog.Recover(buildTime, FUID, Fversion, args[1:])
	BuildTime = buildTime
	BuildHost = buildHost
	GoVersion = goVersion
	Fpayload = payload
	if payload != nil && payload.Origin != "" {
		OQtechType = payload.Origin
	}
	Fmsg = ""
	err := rootCmd.Execute()
	stderr := ""
	if err != nil && len(args) != 1 {
		stderr = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", Ferrid)
		if stderr != "" {
			l := log.New(os.Stderr, "", 0)
			l.Println(stderr)
		}
		os.Exit(1)
	}
	if Fmsg != "" {
		frune := '{'
		for _, c := range Fmsg {
			frune = c
			break
		}
		if Frstripln {
			Fmsg = strings.TrimRight(Fmsg, "\n\r")
		}
		if !Fsilent {
			if Fstdout == "" || Ftransported {
				if frune == '[' || frune == '{' {
					fmt.Println(Fmsg)
				} else {
					fmt.Print(Fmsg)
				}
				return
			}
			f, err := os.Create(Fstdout)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			w := bufio.NewWriter(f)
			if frune == '[' || frune == '{' {
				fmt.Fprintln(w, Fmsg)
			} else {
				fmt.Fprint(w, Fmsg)
			}
			err = w.Flush()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func checkCwd(cwd string) (dir string, err error) {
	if Ftransported {
		cwd = ""
	}
	defer func() { dir, _ = qfs.AbsPath(dir) }()
	dir = cwd
	if cwd == "" {
		dir, err = os.Getwd()
		if err != nil {
			e := &qerror.QError{
				Ref: []string{errRoot + "cwd"},
				Msg: []string{"Cannot determine `cwd`"},
			}
			err = qerror.QErrorTune(err, e)
			return
		}
	}
	if !qfs.Exists(dir) || !qfs.IsDir(dir) {
		err = &qerror.QError{
			Ref: []string{errRoot + "cwd.isdir"},
			Msg: []string{"Path does not exists or is not a directory"},
		}
		return
	}
	return
}

func checkUID(uid string) (usr string) {
	usr = uid
	if usr == "" {
		usr = qregistry.Registry["qtechng-user"]
	}
	if usr == "" {
		pusr, e := user.Current()
		if e == nil {
			usr = pusr.Username
			if usr == "root" {
				usr = ""
			}
		}
	}

	if usr == "" {
		usr = "root"
	}
	return
}

func fillVersion() {
	if Fversion != "" {
		return
	}
	dir := new(qclient.Dir)
	dir.Dir = Fcwd
	files := dir.Repository()
	if len(files) == 1 {
		for r := range files {
			Fversion = r
			break
		}
	}
	switch {
	case strings.ContainsRune(QtechType, 'P'):
		Fversion = qregistry.Registry["brocade-release"]
	case strings.ContainsRune(QtechType, 'W'):
		Fversion = qregistry.Registry["qtechng-version"]
	default:
		Fversion = "0.00"
	}
}

func fillQdir() {
	if Fqdir != "" || Ftransported {
		return
	}
	fillVersion()

	dir := new(qclient.Dir)
	dir.Dir = Fcwd
	files := dir.Repository()
	qdirs := make(map[string][]string)
	alldirs := make(map[string]bool)
	for r, mdir := range files {
		_, ok := qdirs[r]
		if !ok {
			qdirs[r] = make([]string, 0)
		}
		for qdir := range mdir {
			qdirs[r] = append(qdirs[r], qdir)
			alldirs[qdir] = true
		}
	}
	if len(alldirs) == 1 {
		for r := range alldirs {
			Fqdir = r
			return
		}
	}
	if Fversion != "" {
		qr, ok := qdirs[Fversion]
		if ok && len(qr) == 1 {
			Fqdir = qr[0]
			return
		}
	}

	if strings.ContainsRune(QtechType, 'W') {
		bdir := qregistry.Registry["qtechng-work-dir"]
		rel, e := filepath.Rel(bdir, Fcwd)
		if e == nil && !strings.HasPrefix(rel, "..") && strings.HasSuffix(Fcwd, rel) {
			rel = filepath.ToSlash(rel)
			if strings.HasPrefix(rel, "./") {
				if rel == "./" {
					rel = ""
				} else {
					rel = rel[2:]
				}
			}
			rel = strings.Trim(rel, "/")
			rel = "/" + rel
		} else {
			rel = ""
		}
		Fqdir = rel
	}
}

func nameCmd(cmd *cobra.Command) string {
	name := strings.SplitN(cmd.Use, " ", 2)[0] + "."
	cmd.VisitParents(func(c *cobra.Command) { name = strings.SplitN(c.Use, " ", 1)[0] + "." + name })
	return name
}
