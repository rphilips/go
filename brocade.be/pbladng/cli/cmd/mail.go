package cmd

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
	ptools "brocade.be/pbladng/lib/tools"
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "mail `gopblad`",
	Long:  "mail `gopblad`",

	Example: `gopblad mail myfile.pb`,
	RunE:    mail,
}

var Fdironly = false

func init() {
	mailCmd.PersistentFlags().StringVar(&Fdocty, "docty", "", "document type: doc, docx, pdf, odt")
	mailCmd.PersistentFlags().BoolVar(&Fdironly, "maildir-only", false, "do not send the mail, create the mail directory")
	rootCmd.AddCommand(mailCmd)
}

func mail(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		for _, m := range pregistry.Registry["edition"].(map[string]any)["receivers"].([]any) {
			a := m.(string)
			a = strings.TrimSpace(a)
			if a == "" {
				continue
			}
			if !strings.Contains(a, "@") {
				return fmt.Errorf("`%s` is not a valid mail address", a)
			}
			args = append(args, a)
		}
	}
	if len(args) == 0 {
		return fmt.Errorf("no mail addresses specified")
	}

	halewijn := false
	mhalewijn := pregistry.Registry["edition"].(map[string]any)["halewijn"].(string)

	for _, m := range args {
		halewijn = m == mhalewijn
		if halewijn {
			break
		}
	}

	source := pfs.FName("workspace/parochieblad.ed")
	if Fdebug {
		Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
		source = filepath.Join(Fcwd, "parochieblad.ed")
	}
	if Fdocty == "" {
		Fdocty = "doc"
	}

	doc, target, err := makeDoc(source, Fdocty)
	if err != nil {
		return err
	}

	attach := make([]string, 0)
	pcode := pregistry.Registry["pcode"].(string)

	maildir := filepath.Join(filepath.Dir(pfs.FName("workspace/parochieblad.ed")), "mail")
	bfs.Rmpath(maildir)
	bfs.MkdirAll(maildir, "process")

	docname := filepath.Join(maildir, fmt.Sprintf("P%s-%02d.doc", pcode, doc.Week))
	bfs.CopyFile(target, docname, "process", false)
	attach = append(attach, docname)

	for _, c := range doc.Chapters {
		for _, t := range c.Topics {
			for _, img := range t.Images {
				imgname := filepath.Join(maildir, fmt.Sprintf("F%s%s%02d.jpg", pcode, img.Letter, doc.Week))
				err = pstructure.ReduceSize(img.Fname, -1)
				if err != nil {
					return err
				}
				bfs.CopyFile(img.Fname, imgname, "process", false)
				attach = append(attach, imgname)
			}
		}
	}

	subject := doc.Title()

	sort.Strings(attach)

	fmt.Println("Subject:", subject)
	fmt.Println("Dir    :", maildir)
	fmt.Println("Attach :")
	for _, f := range attach {
		fmt.Println("   ", f)
	}
	fmt.Println("To     :")
	for _, m := range args {
		fmt.Println("   ", m)
	}
	if Fdironly {
		return err
	}

	ok := ptools.YesNo("Send mail ?")

	if !ok {
		return err
	}
	text := doc.MailText()

	if true {
		err = bmail.Send(args, nil, nil, subject, text, "", attach, "")
	}
	halewijn = true
	if err == nil && halewijn {
		now := time.Now()
		doc.Mailed = &now
		err = bfs.Store(source, doc.String(), "process")
		if err == nil {
			fmt.Println("Mail sent!")
		}
		err = doc.Archive()
		workbase(doc)
	}

	return err
}

func workbase(doc *pstructure.Document) {
	names := doc.Names()
	if len(names) == 0 {
		return
	}
	pnames := pfs.FName("support/names.txt")
	data, _ := bfs.Fetch(pnames)
	sdata := strings.SplitN(string(data), "\n", -1)

	m := make(map[string]bool)
	lines := make([]string, 0, len(sdata)+len(names))
	ok := false
	for _, x := range sdata {
		x = strings.TrimSpace(x)
		if x == "" {
			continue
		}
		lines = append(lines, x)
		m[x] = true
	}
	for _, name := range names {
		if m[name] {
			continue
		}
		m[name] = true
		lines = append(lines, name)
		ok = true
	}
	if !ok {
		return
	}
	sort.Strings(lines)
	lines = append(lines, "")
	bfs.Store(pnames, strings.Join(lines, "\n"), "process")
}
