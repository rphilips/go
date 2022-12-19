package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	pdocument "brocade.be/pbladng/lib/document"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
)

var distributeCmd = &cobra.Command{
	Use:   "distribute",
	Short: "Distribute `gopblad`",
	Long:  "Distribute `gopblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `gopblad distribute myfile.pb`,
	RunE:    distribute,
}

func init() {
	rootCmd.AddCommand(distributeCmd)
}

func distribute(cmd *cobra.Command, args []string) error {

	d, ok := pregistry.Registry["distribute-dir"]

	if !ok || d.(string) == "" {
		d = pfs.FName("workspace")
	}
	mails := make([]string, 0)
	dir := d.(string)
	rex := regexp.MustCompile("^[0-9]{4}-[0-9]{2}$")
	id := ""
	for _, base := range []string{"parochieblad.docx", "nazareth.docx"} {
		fname := filepath.Join(dir, base)
		if !bfs.Exists(fname) {
			continue
		}
		year, week, err := pdocument.DocRef(dir)
		if err != nil {
			return err
		}
		id = fmt.Sprintf("%d-%02d", year, week)
		correspondents := pregistry.Registry["correspondents"].(map[string]any)

		for _, x := range correspondents {
			mail := x.(map[string]any)["mail"].(string)
			mails = append(mails, mail)
			cdir := x.(map[string]any)["dir"].(string)
			tdir := filepath.Join(cdir, id)
			bfs.Mkdir(tdir, "process")
			target := filepath.Join(tdir, "parochieblad.docx")
			if !bfs.Exists(target) {
				err := bfs.CopyFile(fname, target, "", false)
				if err != nil {
					return err
				}
			}
			if !bfs.IsFile(target) {
				err := fmt.Errorf("could not copy to %s", target)
				return err
			}

			_, dirs, err := bfs.FilesDirs(cdir)
			if err != nil {
				return err
			}
			deldirs := make([]string, 0)
			for _, d := range dirs {
				if !rex.MatchString(d.Name()) {
					continue
				}
				if d.Name() >= id {
					continue
				}
				deldirs = append(deldirs, d.Name())
			}
			sort.Strings(deldirs)
			if len(deldirs) < 5 {
				continue
			}
			deldirs = deldirs[:len(deldirs)-4]
			for _, d := range deldirs {
				bfs.Rmpath(filepath.Join(cdir, d))
			}
		}

	}
	if len(mails) != 0 {
		// "year": {
		//     "2022": {
		//         "max": "52",
		//         "thursday": "",
		//         "holiday": ""
		//     },
		//     "2023": {
		//         "max": "52",
		//         "thursday": "15,18,22,33,52",
		//         "holiday": "26/27/28,29/30/31"
		//     }

		year, week, _ := strings.Cut(id, "-")

		msgcode := ""
		specials := pregistry.Registry["year"].(map[string]any)
		_, ok := specials[year]
		if ok {
			specials := specials[year].(map[string]any)
			x, ok := specials["thursday"]
			if ok {
				weeks := "," + x.(string) + ","
				if strings.Contains(weeks, ","+week+",") {
					msgcode = "thursday"
				}
			}
			if msgcode == "" {
				x, ok := specials["holiday"]
				if ok {
					weeks := strings.ReplaceAll(","+x.(string)+",", "/", ",")
					if strings.Contains(weeks, ","+week+",") {
						msgcode = "holiday"
					}
				}
			}
		}

		subject := pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["subject"].(string)
		msg := ""
		if msgcode != "" {
			msg = pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["subject-"+msgcode].(string)
		}
		fmt.Println(strings.TrimLeft(msg, ": "))
		subject = strings.ReplaceAll(subject, "{id}", id)
		subject = strings.ReplaceAll(subject, "{msg}", msg)

		body := pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["body"].(string)
		msg = ""
		if msgcode != "" {
			msg = pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["body-"+msgcode].(string)
		}
		body = strings.ReplaceAll(body, "{id}", id)
		body = strings.ReplaceAll(body, "{msg}", msg)
		err := bmail.Send(mails, nil, nil, subject, body, "", nil)
		if err != nil {
			return err
		}

	}

	return nil
}
