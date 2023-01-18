package cmd

import (
	"bytes"
	"fmt"
	"html"
	"os/exec"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	vregistry "brocade.be/vchess/lib/registry"
	vstrings "brocade.be/vchess/lib/strings"
	vstructure "brocade.be/vchess/lib/structure"
	"github.com/spf13/cobra"
)

var Fhtml = false
var Fpdf = false
var Fmail = false
var Fname = ""

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Information print `vchess`",
	Long:  `Version and build time printrmation print the vchess executable`,

	Args:    cobra.MaximumNArgs(1),
	Example: `vchess print`,
	RunE:    print,
}

func init() {
	printCmd.PersistentFlags().BoolVar(&Fhtml, "html", false, "HTML output")
	printCmd.PersistentFlags().BoolVar(&Fpdf, "pdf", false, "PDF output")
	printCmd.PersistentFlags().BoolVar(&Fmail, "mail", false, "Mail output")
	printCmd.PersistentFlags().StringVar(&Fname, "name", "", "filter op teamname")
	rootCmd.AddCommand(printCmd)
}

func print(cmd *cobra.Command, args []string) (err error) {

	if len(args) == 0 {
		args = append(args, "R0")
	}
	if !Fpdf && !Fmail {
		Fhtml = true
	}
	if Fmail {
		Fpdf = true
	}
	r := strings.TrimSpace(strings.ReplaceAll(args[0], "R", ""))
	ri, err := strconv.Atoi(r)
	if err != nil {
		return
	}
	season := new(vstructure.Season)
	season.Init(nil)
	round, err := season.Round(ri)
	if err != nil {
		return
	}

	allduels := make([]*vstructure.Duel, 0)

	homeplay := true

	if Fname == "" {
		allduels = round.Duels
	} else {
		for _, duel := range round.Duels {
			teamh := duel.Home
			teamr := duel.Remote
			if !strings.Contains(teamh.Name, Fname) && !strings.Contains(teamr.Name, Fname) {
				continue
			}
			allduels = append(allduels, duel)
		}
	}

	for _, duel := range allduels {
		teamh := duel.Home
		teamr := duel.Remote
		if teamr.VSL {
			homeplay = false
		}
		fmt.Println("Seizoen :", season)
		fmt.Println("Ronde   :", round)
		fmt.Println("Afdeling:", teamh.Division)
		fmt.Println("Datum   :", round.Date.Format(time.RFC3339)[:10])
		fmt.Println("Teams   :", teamh.String()+" ("+teamh.Club.Stamno+") vs. "+teamr.String()+" ("+teamr.Club.Stamno+")")
		fmt.Println()

		maxn := 0
		maxs := 0

		for _, match := range duel.Match {
			p := match.VSL
			q := match.Other
			if len(p.Name) > maxn {
				maxn = len(p.Name)
			}
			if len(q.Name) > maxn {
				maxn = len(q.Name)
			}
			if len(p.Stamno) > maxs {
				maxs = len(p.Stamno)
			}
			if len(q.Stamno) > maxs {
				maxs = len(q.Stamno)
			}
		}
		frame := "%2d.  %{maxs}s  %{maxn}s       %-{maxn}s  %-{maxs}s\n"
		frame = strings.ReplaceAll(frame, "{maxn}", strconv.Itoa(maxn))
		frame = strings.ReplaceAll(frame, "{maxs}", strconv.Itoa(maxs))

		for i, match := range duel.Match {
			p := match.VSL
			q := match.Other
			if duel.Remote.VSL {
				p = match.Other
				q = match.VSL
			}
			fmt.Printf(frame, i+1, p.Stamno, p.Name, q.Name, q.Stamno)

		}

		fmt.Print("\n===\n\n")

	}

	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)
	buffer.WriteString(fmt.Sprintf(`<!DOCTYPE html>
		<html lang="nl">
		<meta charset="UTF-8">
		<title>%s: %s</title>
		<style>
		table.club,
		tr.club,
		tr.club > td
		{
			font-size: x-large;
			vertical-align: top;

		}
		table.score,
		tr.score,
		tr.score > th,
		tr.score > td
		{
	        padding: 10px;
	        border: 1px solid black;
	        border-collapse: collapse;
	      }
		a {
			outline: none;
			text-decoration: none;
			padding: 2px 1px 0;
		}

		a:link {
			color: blue;
			cursor: pointer;
		}

		a:visited {
			color: blue;
			cursor: pointer;
		}

		a:focus {
		}

		a:hover {
		}

		a:active {
		}
		</style>
		<script src=""></script>
		<body>`, season, round))

	escape := html.EscapeString
	for i, duel := range allduels {
		if i != 0 {
			buffer.WriteString(`<p style="page-break-after: always;">&#160;</p>`)
		}
		teamh := duel.Home
		teamr := duel.Remote

		buffer.WriteString(fmt.Sprintf(`<table>
	<tr><td>Seizoen</td><td>%s</td></tr>
	<tr><td>Ronde</td><td>%s</td></tr>
	<tr><td>Afdeling</td><td>%s</td></tr>
	<tr><td>Datum</td><td>%s</td></tr>
	<tr><td>Teams</td><td>%s</td></tr>
	</table>`, season, strings.ReplaceAll(round.String(), "R", ""), teamh.Division, round.Date.Format(time.RFC3339)[:10], `<a href="`+teamh.Club.URL()+`" target="_blank">`+escape(teamh.Name+" ("+teamh.Club.Stamno+")")+"</a> vs. "+`<a href="`+teamr.Club.URL()+`" target="_blank">`+escape(teamr.Name)+" ("+teamr.Club.Stamno+")") + "</a>")

		maxn := 0
		maxs := 0

		for _, match := range duel.Match {
			p := match.VSL
			q := match.Other
			if len(p.Name) > maxn {
				maxn = len(p.Name)
			}
			if len(q.Name) > maxn {
				maxn = len(q.Name)
			}
			if len(p.Stamno) > maxs {
				maxs = len(p.Stamno)
			}
			if len(q.Stamno) > maxs {
				maxs = len(q.Stamno)
			}
		}
		if maxn < len("Frederik Van De Casteele") {
			maxn = len("Frederik Van De Casteele")
		}

		frame := `<tr class="score"><td align="right">%d</td><td  align="right" style="min-width:{maxs}ch;">%s</td><td  align="right" style="min-width:{maxn}ch;">%s</td><td align="center" style="min-width:3em;">&#160;</td><td align="left" style="min-width:{maxn}ch;">%s</td><td align="left" style="min-width:{maxs}ch;">%s</td></tr>`
		frame = strings.ReplaceAll(frame, "{maxn}", strconv.Itoa(maxn))
		frame = strings.ReplaceAll(frame, "{maxs}", strconv.Itoa(maxs))

		buffer.WriteString(`<p>&#160;</p><table class="score">`)
		buffer.WriteString(fmt.Sprintf(`<tr class="score"><th align="right" >Bord</th><th align="center">Stamnr. <br />%s</th><th align="right">Speler <br />%s</th><th align="center">Score</th><th align="left">Speler <br />%s</th><th align="left">Stamnr. <br />%s</th></tr>`, escape(teamh.Name), escape(teamh.Name), escape(teamr.Name), escape(teamr.Name)))

		for i, match := range duel.Match {
			p := match.VSL
			q := match.Other
			if duel.Remote.VSL {
				p = match.Other
				q = match.VSL
			}
			buffer.WriteString(fmt.Sprintf(frame, i+1, p.Stamno, escape(p.Name), escape(q.Name), q.Stamno))
		}
		buffer.WriteString(`</table><p>&#160;</p>`)
		rules := vregistry.Registry["season"].(map[string]any)[season.String()].(map[string]any)["print-rules"].(string)
		buffer.WriteString(rules)
		if !teamh.VSL {
			club := teamh.Club
			club.Load(round.String())
			buffer.WriteString(`<div><table class="club">`)
			buffer.WriteString(`<tr class="club"><td>Lokaal: </td><td>` + escape(club.Site) + "</td></tr>")
			buffer.WriteString(`<tr class="club"><td>Adres: </td><td>` + strings.Join(club.Address, "<br />") + "</td></tr>")
			buffer.WriteString(`<tr class="club"><td>Contact: </td><td>` + strings.Join(club.Contact, "<br />") + "</td></tr>")
			buffer.WriteString(`</table></div>`)
		}
	}

	buffer.WriteString(`</body>
	</html>`)

	mode := "pdf"
	if Fhtml {
		mode = "html"
	}
	outputfile := season.OutputFile(round.String(), "html", Fname)

	bfs.Store(outputfile, buffer.Bytes(), "")

	pdffile := ""
	if mode == "pdf" {
		target := strings.TrimSuffix(outputfile, ".html") + ".pdf"
		pdffile = target
		aconvertor := vregistry.Registry["convert"].(map[string]any)["html2pdf"].([]any)
		convertor := make([]string, 0)
		for _, piece := range aconvertor {
			convertor = append(convertor, strings.ReplaceAll(strings.ReplaceAll(piece.(string), "{source}", outputfile), "{target}", target))
		}
		ccmd := exec.Command(convertor[0], convertor[1:]...)
		err := ccmd.Run()
		if err != nil {
			panic(err)
		}
		outputfile = target
	}

	fmt.Println("\n\n" + outputfile)

	aviewer := vregistry.Registry["viewer"].(map[string]any)["pdf"].([]any)

	if Fhtml {
		aviewer = vregistry.Registry["viewer"].(map[string]any)["html"].([]any)
	}
	viewer := make([]string, 0)

	for _, piece := range aviewer {
		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", outputfile))
	}
	vcmd := exec.Command(viewer[0], viewer[1:]...)
	err = vcmd.Start()
	if err != nil {
		panic(err)
	}
	if Fmail && vstrings.YesNo("Mail the print ?") {
		return mailprint(season.String(), round.String(), pdffile, homeplay)
	}

	return nil

}

func mailprint(season string, round string, pdffile string, home bool) (err error) {
	round = strings.ReplaceAll(round, "R", "")
	to := make([]string, 0)
	cc := make([]string, 0)
	bcc := make([]string, 0)
	mail := vregistry.Registry["season"].(map[string]any)[season].(map[string]any)["mail-print"].(map[string]any)

	subject := mail["subject"].(string)
	subject = strings.ReplaceAll(subject, "{season}", season)
	subject = strings.ReplaceAll(subject, "{round}", round)
	for _, d := range mail["to"].([]any) {
		to = append(to, d.(string))
	}
	for _, d := range mail["cc"].([]any) {
		cc = append(to, d.(string))
	}
	for _, d := range mail["bcc"].([]any) {
		bcc = append(to, d.(string))
	}
	html := mail["content-home"].([]any)
	if !home {
		html = mail["content-remote"].([]any)
	}
	body := ""
	for _, h := range html {
		body += h.(string) + "<br />\n"
	}

	body = strings.ReplaceAll(body, "{season}", season)
	body = strings.ReplaceAll(body, "{round}", round)

	err = bmail.Send(to, cc, bcc, subject, "", string(body), []string{pdffile}, "")
	return err
}
