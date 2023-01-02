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
	bstrings "brocade.be/base/strings"
	pdocument "brocade.be/pbladng/lib/document"
	pfs "brocade.be/pbladng/lib/fs"
	plog "brocade.be/pbladng/lib/log"
	pnext "brocade.be/pbladng/lib/next"
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
	year, week, _, err := pdocument.DocRef(dir)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%d-%02d", year, week)
	value, _ := plog.GetMark("distribute")
	if value != "" && value >= id {
		return nil
	}

	rex := regexp.MustCompile("^[0-9]{4}-[0-9]{2}$")
	for _, base := range []string{"parochieblad.docx", "nazareth.docx"} {
		fname := filepath.Join(dir, base)
		if !bfs.Exists(fname) {
			continue
		}
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
				fmt.Println("→", target)
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
		//mails = []string{"richard.philips@gmail.com"}
		//id = "2023-27"
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

		_, msgcode := pnext.Special(id)

		if msgcode == "holiday" {
			return nil
		}

		subject := pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["subject"].(string)
		if msgcode != "" {
			subject = pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["subject-"+msgcode].(string)
		}
		subject = strings.ReplaceAll(subject, "{id}", id)

		body := pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["body"].(string)
		if msgcode != "" {
			body = pregistry.Registry["mails"].(map[string]any)["correspondents"].(map[string]any)["body-"+msgcode].(string)
		}
		body = strings.ReplaceAll(body, "{id}", id)
		err := bmail.Send(mails, nil, nil, subject, body, "", nil)
		if err != nil {
			return err
		}
		fmt.Println(bstrings.JSON(mails))
		plog.SetMark("distribute", id)

		interested := pregistry.Registry["mails"].(map[string]any)["interested"].(map[string]any)
		amails, ok := interested["to"]
		if !ok {
			return nil
		}
		mails = amails.([]string)
		if len(mails) == 0 {
			return nil
		}
		asubject, ok := interested["subject-"+msgcode]
		if !ok {
			return nil
		}
		subject = asubject.(string)
		if subject == "" {
			return nil
		}
		abody, ok := interested["body-"+msgcode]
		if !ok {
			return nil
		}
		body = abody.(string)
		if body == "" {
			return nil
		}
		subject = strings.ReplaceAll(subject, "{id}", id)
		body = strings.ReplaceAll(body, "{id}", id)
		err = bmail.Send(mails, nil, nil, subject, body, "", nil)
		if err != nil {
			return err
		}
		fmt.Println(bstrings.JSON(mails))
	}

	return nil
}
