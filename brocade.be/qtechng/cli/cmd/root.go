package cmd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qserver "brocade.be/qtechng/lib/server"
	"github.com/spf13/cobra"
	"github.com/spyzhov/ajson"
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

var cfgFile string

// Fcwd Current working directory
var Fcwd string // current working directory

//Fenv environment variables
var Fenv []string

// FUID userid
var FUID string

// Fpayload pointer to payload information
var Fpayload *qclient.Payload

// Fcargo pointer to payoff information
var Fcargo *qclient.Cargo = &qclient.Cargo{}

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

//Fsmartcaseoff smartcase in search staat off
var Fsmartcaseoff bool

// Fneedle needles to search for
var Fneedle []string

// Fpattern patterns to select on
var Fpattern []string

// Fforce overrules normal behaviour
var Fforce bool // force ?
var finfo bool  // Fakes command and lists all arguments

// Feditor editor used in development
var Feditor string // editor used in development
// Fmsg result as a string
var Fmsg string // result as a string

// Fjq JSONPath
var Fjq string

// Funhex decides if the args are to be unhexed (if starting with `.`)
var Funhex bool

// Fcmpversion version to compare with
var Fcmpversion string

// Ffilesinproject geeft aan of de bestanden in het project ook dienen te worden  geslecteerd
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

// Finstallref reference to the installation
var Finstallref string

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
	rootCmd.PersistentFlags().MarkHidden("transported")
	rootCmd.PersistentFlags().MarkHidden("uid")

	rootCmd.PersistentFlags().StringVar(&Fcwd, "cwd", "", "Working directory")
	rootCmd.PersistentFlags().BoolVar(&Funhex, "unhex", false, "Unhexify the arguments starting with `.`")
	rootCmd.PersistentFlags().StringVar(&Feditor, "editor", "", "editor name")
	rootCmd.PersistentFlags().StringVar(&Fjq, "jsonpath", "", "JSONpath")
	rootCmd.PersistentFlags().BoolVar(&Fsilent, "quiet", false, "Silent the output")
	rootCmd.PersistentFlags().StringSliceVar(&Fenv, "env", []string{}, "Environment variable KEY=VALUE")
	// rootCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	// rootCmd.PersistentFlags().StringVar(&Fproject, "project", "", "project to work with")
	// rootCmd.PersistentFlags().StringVar(&Fqdir, "qdir", "", "qpath of a directory under a project")
	// rootCmd.PersistentFlags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern")
	// rootCmd.PersistentFlags().StringSliceVar(&Fneedle, "needle", []string{}, "Posix glob pattern")
	// rootCmd.PersistentFlags().BoolVar(&Fforce, "force", false, "with force")
	// rootCmd.PersistentFlags().BoolVar(&Fremote, "remote", false, "execute on the remote server")
	// rootCmd.PersistentFlags().BoolVar(&Fperline, "perline", false, "searches per line")
	// rootCmd.PersistentFlags().BoolVar(&finfo, "info", false, "lists arguments, flags and global values")
	// rootCmd.PersistentFlags().BoolVar(&Frecurse, "recurse", false, "recursively walks through directory and subdirectories")
	// rootCmd.PersistentFlags().BoolVar(&Fregexp, "regexp", false, "searches as a regular expression")
	// rootCmd.PersistentFlags().BoolVar(&Ftolower, "tolower", false, "transforms to lowercase")
	// rootCmd.PersistentFlags().BoolVar(&Fsmartcaseoff, "smartcaseoff", false, "transforms with smartcase")
	// rootCmd.PersistentFlags().BoolVar(&Ftransported, "transported", false, "Indicate if comamnd is transported")
	// rootCmd.PersistentFlags().BoolVar(&Ffilesinproject, "neighbours", false, "Indicate if all files in project are selected")

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

	// Jq
	if Fjq != "" && !Ftransported {
		_, err = ajson.ParseJSONPath(Fjq)
		if err != nil {
			err = &qerror.QError{
				Ref: []string{errRoot + "jsonpath"},
				Msg: []string{fmt.Sprintf("JSONpath `" + Fjq + "` error: " + err.Error())},
			}
			return
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
	FUID = checkUID(FUID)

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
			Msg: []string{fmt.Sprintf("Command is not allowed on remote server")},
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

	withforce, _ := cmd.Annotations["with-force"]
	if withforce == "yes" && !Fforce {
		err = &qerror.QError{
			Ref: []string{errRoot + "force"},
			Msg: []string{"Command should be run with --force=true"},
		}
	}

	// root
	withroot, _ := cmd.Annotations["with-root"]
	yes := withroot == "yes" || withroot == "*"
	if !yes && withroot != "" {
		for _, char := range QtechType {
			yes = strings.ContainsRune(allowed, char)
			if yes {
				break
			}
		}
	}
	if yes {
		euid := syscall.Geteuid()
		suid, _ := user.Current()
		if euid != -1 {
			suid, _ = user.LookupId(strconv.Itoa(euid))
		}
		if suid.Username != "root" {
			err = &qerror.QError{
				Ref: []string{errRoot + "root"},
				Msg: []string{fmt.Sprintf("Command should be run as root, current is `%s`", suid.Username)},
			}
		}
	}

	if err != nil {
		return
	}

	return
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(buildTime string, goVersion string, buildHost string, payload *qclient.Payload) {
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
	if err != nil && len(os.Args) != 1 {
		stderr = qerror.ShowError(err)
		if stderr != "" {
			l := log.New(os.Stderr, "", 0)
			l.Println(stderr)
		}
		os.Exit(1)
	}
	if Fmsg != "" && !Fsilent {
		frune := '{'
		for _, c := range Fmsg {
			frune = c
			break
		}
		if Frstripln {
			Fmsg = strings.TrimRight(Fmsg, "\n\r")
		}
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
		switch {
		case strings.ContainsRune(QtechType, 'W'):
			usr = qregistry.Registry["qtechng-user"]
		default:
			usr = ""
		}
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
		usr = "usystem"
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
			return
		}
	}
	Fversion = qregistry.Registry["qtechng-version"]
	Fversion = qserver.Canon(Fversion)

	return
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
			for _, r := range qr {
				Fqdir = r
				return
			}
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
