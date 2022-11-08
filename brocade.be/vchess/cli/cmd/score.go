package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	vicyear "brocade.be/vchess/lib/icyear"
	vregistry "brocade.be/vchess/lib/registry"
	"github.com/spf13/cobra"
)

var scoreCmd = &cobra.Command{
	Use:   "score",
	Short: "Information print `vchess`",
	Long:  `Version and build time printrmation print the vchess executable`,

	Args:    cobra.MaximumNArgs(1),
	Example: `vchess score`,
	RunE:    score,
}

func init() {
	scoreCmd.PersistentFlags().BoolVar(&Fhtml, "html", false, "HTML output")
	scoreCmd.PersistentFlags().BoolVar(&Fpdf, "pdf", false, "PDF output")
	scoreCmd.PersistentFlags().BoolVar(&Fmail, "mail", false, "mail output")
	rootCmd.AddCommand(scoreCmd)
}

func score(cmd *cobra.Command, args []string) error {
	if Fmail {
		Fhtml = false
		Fpdf = true
	}
	last := ""
	if len(args) == 0 {
		for i := 1; i < 20; i++ {
			fname := vicyear.ActiveRound(nil, "R"+strconv.Itoa(i))
			if bfs.Exists(fname) {
				last = "R" + strconv.Itoa(i)
				continue
			}
		}
		args = append(args, last)
	}
	round := strings.Trim(strings.ToUpper(args[0]), "R ")
	_, err := strconv.Atoi(round)
	if err != nil {
		return fmt.Errorf("argument should be a round number")
	}
	round = "R" + round
	fname := vicyear.ActiveRound(nil, round)
	if !bfs.Exists(fname) {
		return fmt.Errorf("no information about `%s`", round)
	}

	teams := vicyear.Teams(nil, nil)
	matches := vicyear.Matches(nil, nil, round)

	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})
	season := ""
	round = ""
	base := vregistry.Registry["club"].(map[string]any)["basename"].(string)
	for i, match := range matches {
		if i != 0 {
			fmt.Println()
		}
		hteam := match.Home
		hnr := vicyear.ClubNummer(nil, hteam.Name)

		rteam := match.Remote
		rnr := vicyear.ClubNummer(nil, rteam.Name)
		round = match.Round
		actives := vicyear.Actives(nil, round)
		season = match.Season
		fmt.Println("Seizoen :", match.Season)
		fmt.Println("Ronde   :", strings.ReplaceAll(round, "R", ""))
		fmt.Println("Afdeling:", hteam.Division)
		fmt.Println("Datum   :", match.Date.Format(time.RFC3339)[:10])
		fmt.Println("Teams   :", hteam.Name+" ("+hnr+")"+" vs. "+rteam.Name+" ("+rnr+")")
		fmt.Println()

		maxn := 0
		maxs := 0
		data := make([][6]string, 0)
		if strings.HasPrefix(hteam.Name, base) {
			duels := actives[hteam.Name]

			for i = 0; i < len(duels); i++ {
				j := strconv.Itoa(i + 1)
				duel := duels[j]
				ph := duel.Home
				pr := duel.Remote
				data = append(data, [6]string{j, ph.Elo, ph.Name, duel.Score, pr.Name, pr.Elo})
				if len(ph.Elo) > maxs {
					maxs = len(ph.Elo)
				}
				if len(pr.Elo) > maxs {
					maxs = len(pr.Elo)
				}
				if len(ph.Name) > maxn {
					maxn = len(ph.Name)
				}
				if len(pr.Name) > maxn {
					maxn = len(pr.Name)
				}
			}
		} else {
			duels := actives[rteam.Name]
			for i = 0; i < len(duels); i++ {
				j := strconv.Itoa(i + 1)
				duel := duels[j]
				ph := duel.Home
				pr := duel.Remote
				if len(ph.Elo) > maxs {
					maxs = len(ph.Elo)
				}
				if len(pr.Elo) > maxs {
					maxs = len(pr.Elo)
				}
				if len(ph.Name) > maxn {
					maxn = len(ph.Name)
				}
				if len(pr.Name) > maxn {
					maxn = len(pr.Name)
				}
				data = append(data, [6]string{j, pr.Elo, pr.Name, duel.Score, ph.Name, ph.Elo})
			}
		}
		frame := "%2s.  %{maxs}s  %{maxn}s  %3s  %-{maxn}s  %-{maxs}s\n"
		frame = strings.ReplaceAll(frame, "{maxs}", strconv.Itoa(maxs))
		frame = strings.ReplaceAll(frame, "{maxn}", strconv.Itoa(maxn))
		for _, line := range data {
			fmt.Printf(frame, line[0], line[1], line[2], line[3], line[4], line[5])
		}
		fmt.Println()
		fmt.Println("----")

	}

	if !Fhtml && !Fpdf {
		return nil
	}
	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)
	buffer.WriteString(fmt.Sprintf(`<!DOCTYPE html>
	<html lang="nl">
	<meta charset="UTF-8">
	<title>%s: %s</title>
	<style>
	.clipboard {cursor: copy;}
	table.score,
	tr.score,
	tr.score > th,
	tr.score > td
	{
        padding: 10px;
        border: 1px solid black;
        border-collapse: collapse;
      }

	table.sum,
	tr.sum,
	tr.sum > th,
	tr.sum > td
	{
		border: none;
		border-bottom: 0;
        padding: 10px;
      }


	</style>
	<script src=""></script>
	<body>`, season, round))

	for i, match := range matches {
		if i != 0 {
			buffer.WriteString(`<p style="page-break-after: always;">&#160;</p>`)
		}
		hteam := match.Home
		hnr := vicyear.ClubNummer(nil, hteam.Name)

		rteam := match.Remote
		rnr := vicyear.ClubNummer(nil, rteam.Name)
		round = match.Round
		actives := vicyear.Actives(nil, round)
		buffer.WriteString(fmt.Sprintf(`<table>
<tr><td>Seizoen</td><td>%s</td></tr>
<tr><td>Ronde</td><td>%s</td></tr>
<tr><td>Afdeling</td><td>%s</td></tr>
<tr><td>Datum</td><td>%s</td></tr>
<tr><td>Teams</td><td>%s</td></tr>
</table>
`, match.Season, strings.ReplaceAll(round, "R", ""), hteam.Division, match.Date.Format(time.RFC3339)[:10], hteam.Name+" ("+hnr+")"+" vs. "+rteam.Name+" ("+rnr+")"))
		maxn := 0
		maxs := 0
		data := make([][8]string, 0)
		sumscore := "0-0"
		reverse := false
		if strings.HasPrefix(hteam.Name, "LANDEGEM") {
			duels := actives[hteam.Name]
			for i = 0; i < len(duels); i++ {
				j := strconv.Itoa(i + 1)
				duel := duels[j]
				ph := duel.Home
				pr := duel.Remote
				sumscore = sum(sumscore, duel.Score)
				if pr.Elo == "0" || pr.Elo == "" {
					pr.Elo = "-"
				}
				if ph.Elo == "0" || ph.Elo == "" {
					ph.Elo = "-"
				}
				data = append(data, [8]string{j, ph.Elo, ph.Name, duel.Score, pr.Name, pr.Elo, ph.Stam, pr.Stam})
				if len(ph.Elo) > maxs {
					maxs = len(ph.Elo)
				}
				if len(pr.Elo) > maxs {
					maxs = len(pr.Elo)
				}
				if len(ph.Name) > maxn {
					maxn = len(ph.Name)
				}
				if len(pr.Name) > maxn {
					maxn = len(pr.Name)
				}
			}
		} else {
			reverse = true
			duels := actives[rteam.Name]
			for i = 0; i < len(duels); i++ {
				j := strconv.Itoa(i + 1)
				duel := duels[j]
				ph := duel.Home
				pr := duel.Remote
				sumscore = sum(sumscore, duel.Score)
				if pr.Elo == "0" || pr.Elo == "" {
					pr.Elo = "-"
				}
				if ph.Elo == "0" || ph.Elo == "" {
					ph.Elo = "-"
				}
				if len(ph.Elo) > maxs {
					maxs = len(ph.Elo)
				}
				if len(pr.Elo) > maxs {
					maxs = len(pr.Elo)
				}
				if len(ph.Name) > maxn {
					maxn = len(ph.Name)
				}
				if len(pr.Name) > maxn {
					maxn = len(pr.Name)
				}
				data = append(data, [8]string{j, pr.Elo, pr.Name, duel.Score, ph.Name, ph.Elo, pr.Stam, ph.Stam})
			}
		}
		escape := html.EscapeString
		buffer.WriteString(`<p>&#160;</p><table class="score">`)
		buffer.WriteString(fmt.Sprintf(`<tr class="score"><th align="right" >Bord</th><th align="center">ELO <br />%s</th><th align="right">Speler <br />%s</th><th align="center">Score</th><th align="left">Speler <br />%s</th><th align="left">ELO <br />%s</th></tr>`, escape(hteam.Name), escape(hteam.Name), escape(rteam.Name), escape(rteam.Name)))

		for _, line := range data {
			escape := func(nr int) string { return html.EscapeString(line[nr]) }
			buffer.WriteString(fmt.Sprintf(`<tr class="score"><td align="right">%s</td><td class="clipboard" data-clip="%s" align="right" style="min-width:%dex;">%s</td><td class="clipboard"  align="right" style="min-width:%dex;">%s</td><td class="clipboard" align="center" style="min-width:3em;">%s</td><td class="clipboard" align="left" style="min-width:%dex;">%s</td><td data-clip="%s" class="clipboard" align="left" style="min-width:%dex;">%s</td></tr>`, escape(0), escape(6), maxs, escape(1), maxn, escape(2), escape(3), maxn, escape(4), escape(7), maxs, escape(5)))
		}
		buffer.WriteString(fmt.Sprintf(`<tr class="sum"><td align="right"></td><td align="right" ></td><td align="right"></td><td class="clipboard" align="center" style="min-width:5em;"><b>%s</b></td><td></td><td></td></tr>`, color(sumscore, reverse)))
		buffer.WriteString(`</table>`)
	}
	buffer.WriteString(`<script>
function toClipboard() {
	var copyText = this.getAttribute('data-clip');
	if (!copyText) {
	    copyText = this.innerHTML;
	}
	copyText = copyText.trim();
	navigator.clipboard.writeText(copyText).then(() => {
		/* clipboard successfully set */
		}, () => {
		/* clipboard write failed */
		});
	}
var elements = document.getElementsByClassName("clipboard");
for (let i in elements) {
	let elm = elements.item(i);
	elm.addEventListener('click', toClipboard);
}
</script>
`)

	buffer.WriteString(`</body>
</html>`)

	mode := "pdf"
	if Fhtml {
		mode = "html"
	}
	outputfile := vicyear.OutputFile(nil, round, "html")
	htmlfile := outputfile
	pdffile := ""

	bfs.Store(outputfile, buffer.Bytes(), "")

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
	if Fmail && YesNo("Mail the score ?") {
		return mailscore(season, round, htmlfile, pdffile)
	}

	return nil

}

func sum(sumscore string, score string) string {
	score = strings.ReplaceAll(score, " ", "")
	x1, x2, _ := strings.Cut(score, "-")
	s1, s2, _ := strings.Cut(sumscore, "-")
	half := "½"
	add1 := 0
	add2 := 0
	if strings.Contains(s1, half) {
		add1 += 1
		s1 = strings.ReplaceAll(s1, half, "")
	}
	if strings.Contains(s2, half) {
		add2 += 1
		s2 = strings.ReplaceAll(s2, half, "")
	}
	if x1 == x2 {
		add1 += 1
		add2 += 1
	} else {
		if x2 != "0" {
			add2 += 2
		}
		if x1 != "0" {
			add1 += 2
		}
	}
	is1, _ := strconv.Atoi(s1)
	is2, _ := strconv.Atoi(s2)
	is1 *= 2
	is2 *= 2
	is1 += add1
	is2 += add2

	if is1%2 == 0 {
		s1 = strconv.Itoa(is1 / 2)
	} else {
		s1 = strconv.Itoa((is1-1)/2) + half
	}
	if is2%2 == 0 {
		s2 = strconv.Itoa(is2 / 2)
	} else {
		s2 = strconv.Itoa((is2-1)/2) + half
	}
	if s1 == ("0" + half) {
		s1 = half
	}
	if s2 == ("0" + half) {
		s2 = half
	}

	return s1 + "-" + s2

}

func color(sumscore string, reverse bool) string {
	s1, s2, _ := strings.Cut(sumscore, "-")

	if s1 == s2 {
		return `<span style="color: blue;">` + sumscore + "</span>"
	}
	half := "½"
	if s1 == half {
		s1 = "0" + s1
	}
	if s2 == half {
		s2 = "0" + s2
	}
	s1 = strings.ReplaceAll(s1, half, "")
	s2 = strings.ReplaceAll(s2, half, "")
	is1, _ := strconv.Atoi(s1)
	is2, _ := strconv.Atoi(s2)
	if is1 > is2 {
		if !reverse {
			return `<span style="color: green;">` + sumscore + "</span>"
		} else {
			return `<span style="color: red;">` + sumscore + "</span>"
		}
	}
	if is1 < is2 {
		if reverse {
			return `<span style="color: green;">` + sumscore + "</span>"
		} else {
			return `<span style="color: red;">` + sumscore + "</span>"
		}
	}
	return sumscore
}

func YesNo(s string) bool {
	for {
		fmt.Printf("%s [y/n] ", s)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(strings.ToLower(text))
		if strings.HasPrefix(text, "y") {
			return true
		}
		if strings.HasPrefix(text, "n") {
			return false
		}
	}
}

func mailscore(season string, round string, htmlfile string, pdffile string) (err error) {
	to := make([]string, 0)
	cc := make([]string, 0)
	bcc := make([]string, 0)
	mail := vregistry.Registry["season"].(map[string]any)[season].(map[string]any)["mail-score"].(map[string]any)

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
	html, err := bfs.Fetch(htmlfile)
	if err != nil {
		return err
	}

	err = bmail.Send(to, cc, bcc, subject, "", string(html), []string{pdffile})
	return err
}
