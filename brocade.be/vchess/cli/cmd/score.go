package cmd

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"strconv"
	"strings"

	bfs "brocade.be/base/fs"
	bmail "brocade.be/base/gmail"
	vregistry "brocade.be/vchess/lib/registry"
	vstrings "brocade.be/vchess/lib/strings"
	vstructure "brocade.be/vchess/lib/structure"
	"github.com/spf13/cobra"
)

var Fcsv bool

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
	scoreCmd.PersistentFlags().BoolVar(&Fmail, "csv", false, "csv output")
	rootCmd.AddCommand(scoreCmd)
}

func score(cmd *cobra.Command, args []string) (err error) {
	season := new(vstructure.Season)
	season.Init(nil)
	_, lastround, results, err := season.Results()
	if err != nil {
		return err
	}
	if lastround == 0 {
		return fmt.Errorf("no scores found")
	}
	if len(args) == 0 {
		args = append(args, "R"+strconv.Itoa(lastround))
	}
	sround := strings.Trim(strings.ToUpper(args[0]), "R ")
	nr, err := strconv.Atoi(sround)
	if err != nil {
		return fmt.Errorf("argument should be a round number")
	}
	fround := season.FName("active-round")
	f := strings.ReplaceAll(fround, "{round}", sround)
	_, e := os.ReadFile(f)
	if e != nil {
		return fmt.Errorf("no information about `R%s`", sround)
	}
	if nr > lastround {
		return fmt.Errorf("no information yet about `R%s`", sround)
	}
	escape := html.EscapeString
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
			<body>`, season, sround))

	oldhometeam := ""
	oldresult := new(vstructure.Result)
	wframe := `<tr class="score"><td align="right">%s</td><td class="clipboard" data-clip="%s" align="right">%s</td><td class="clipboard"  data-clip="%s" align="right">%s</td><td class="clipboard" align="center">%s</td><td class="clipboard" data-clip="%s" align="left">%s</td><td data-clip="%s" class="clipboard" align="left">%s</td></tr>`
	for _, result := range results {
		if result.Round != "R"+sround {
			continue
		}
		if oldhometeam != result.TeamhName {
			if oldhometeam != "" {
				buffer.WriteString(fmt.Sprintf(`<tr class="sum"><td align="right"></td><td align="right" ></td><td align="right"></td><td class="clipboard" align="center"><b>%s</b></td><td></td><td></td></tr>`, color(oldresult.TotalScore, oldresult.VSLHome == "0")))
				buffer.WriteString(`</table>`)
				fmt.Println("\n----")
				buffer.WriteString(`<p style="page-break-after: always;">&#160;</p>`)

			}

			oldhometeam = result.TeamhName
			fmt.Println("Seizoen :", result.Season)
			fmt.Println("Ronde   :", sround)
			fmt.Println("Afdeling:", result.Division)
			fmt.Println("Datum   :", result.Date)
			fmt.Println("Teams   :", result.TeamhName+" ("+result.TeamhClubno+")"+" vs. "+result.TeamrName+" ("+result.TeamrClubno+")")
			fmt.Println("Score   :", result.TotalScore)
			fmt.Println()
			buffer.WriteString(fmt.Sprintf(`
<table>
<tr><td>Seizoen</td><td>%s</td></tr>
<tr><td>Ronde</td><td>%s</td></tr>
<tr><td>Afdeling</td><td>%s</td></tr>
<tr><td>Datum</td><td>%s</td></tr>
<tr><td>Teams</td><td>%s</td></tr>
<tr><td>Score</td><td>%s</td></tr>
</table>
`, result.Season, sround, result.Division, result.Date, escape(result.TeamhName)+" ("+result.TeamhClubno+")"+" vs. "+escape(result.TeamrName)+" ("+result.TeamrClubno+")", result.TotalScore))
			buffer.WriteString(`<p>&#160;</p><table class="score">`)
			buffer.WriteString(fmt.Sprintf(`<tr class="score"><th align="right" >Bord</th><th align="center">ELO <br />%s</th><th align="right">Speler <br />%s</th><th align="center">Score</th><th align="left">Speler <br />%s</th><th align="left">ELO <br />%s</th></tr>`, escape(result.TeamhName), escape(result.TeamhName), escape(result.TeamrName), escape(result.TeamrName)))
		}
		buffer.WriteString(fmt.Sprintf(wframe, result.Board, result.PlayerhStamno, result.PlayerhELO, result.PlayerhStamno, escape(result.PlayerhName), result.Score, result.PlayerrStamno, escape(result.PlayerrName), result.PlayerrStamno, result.PlayerrELO))
		oldresult = result
	}

	if oldresult != nil {
		buffer.WriteString(fmt.Sprintf(`<tr class="sum"><td align="right"></td><td align="right" ></td><td align="right"></td><td class="clipboard" align="center"><b>%s</b></td><td></td><td></td></tr>`, color(oldresult.TotalScore, oldresult.VSLHome == "0")))
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
	outputfile := season.OutputFile(sround, "html", "")
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
	if Fmail && vstrings.YesNo("Mail the score ?") {
		return mailscore(season.String(), sround, htmlfile, pdffile)
	}

	return err
}

// 				if len(ph.Elo) > maxs {
// 					maxs = len(ph.Elo)
// 				}
// 				if len(pr.Elo) > maxs {
// 					maxs = len(pr.Elo)
// 				}
// 				if len(ph.Name) > maxn {
// 					maxn = len(ph.Name)
// 				}
// 				if len(pr.Name) > maxn {
// 					maxn = len(pr.Name)
// 				}

// 				csvscore = append(csvscore, vicyear.CSV{
// 					Season:     match.Season,
// 					Round:      match.Round,
// 					Date:       match.Date.Format(time.RFC3339)[:10],
// 					Division:   hteam.Division,
// 					HomeNr:     hnr,
// 					HomeName:   hteam.Name,
// 					RemoteNr:   rnr,
// 					RemoteName: rteam.Name,
// 					Board:      j,
// 					WhiteNr:    ph.Stam,
// 					WhiteName:  ph.Name,
// 					WhiteElo:   ph.Elo,
// 					WhiteScore: strings.SplitN(duel.Score+"-", "-", -1)[0],
// 					BlackNr:    pr.Stam,
// 					BlackName:  pr.Name,
// 					BlackScore: strings.SplitN(duel.Score+"-", "-", -1)[1],
// 				})

// 			}
// 		} else {
// 			duels := actives[rteam.Name]
// 			for i = 0; i < len(duels); i++ {
// 				j := strconv.Itoa(i + 1)
// 				duel := duels[j]
// 				ph := duel.Home
// 				pr := duel.Remote
// 				if len(ph.Elo) > maxs {
// 					maxs = len(ph.Elo)
// 				}
// 				if len(pr.Elo) > maxs {
// 					maxs = len(pr.Elo)
// 				}
// 				if len(ph.Name) > maxn {
// 					maxn = len(ph.Name)
// 				}
// 				if len(pr.Name) > maxn {
// 					maxn = len(pr.Name)
// 				}
// 				data = append(data, [6]string{j, pr.Elo, pr.Name, duel.Score, ph.Name, ph.Elo})
// 				csvscore = append(csvscore, vicyear.CSV{
// 					Season:     match.Season,
// 					Round:      match.Round,
// 					Date:       match.Date.Format(time.RFC3339)[:10],
// 					Division:   hteam.Division,
// 					HomeNr:     hnr,
// 					HomeName:   hteam.Name,
// 					RemoteNr:   rnr,
// 					RemoteName: rteam.Name,
// 					Board:      j,
// 					WhiteNr:    ph.Stam,
// 					WhiteName:  ph.Name,
// 					WhiteElo:   ph.Elo,
// 					WhiteScore: strings.SplitN(duel.Score+"-", "-", -1)[0],
// 					BlackNr:    pr.Stam,
// 					BlackName:  pr.Name,
// 					BlackScore: strings.SplitN(duel.Score+"-", "-", -1)[1],
// 				})
// 			}
// 		}
// 		frame := "%2s.  %{maxs}s  %{maxn}s  %3s  %-{maxn}s  %-{maxs}s\n"
// 		frame = strings.ReplaceAll(frame, "{maxs}", strconv.Itoa(maxs))
// 		frame = strings.ReplaceAll(frame, "{maxn}", strconv.Itoa(maxn))
// 		for _, line := range data {
// 			fmt.Printf(frame, line[0], line[1], line[2], line[3], line[4], line[5])
// 		}
// 		fmt.Println()
// 		fmt.Println("----")

// 	}

// 	if !Fhtml && !Fpdf {
// 		return nil
// 	}
// 	buf := make([]byte, 0)
// 	buffer := bytes.NewBuffer(buf)
// 	buffer.WriteString(fmt.Sprintf(`<!DOCTYPE html>
// 	<html lang="nl">
// 	<meta charset="UTF-8">
// 	<title>%s: %s</title>
// 	<style>
// 	.clipboard {cursor: copy;}
// 	table.score,
// 	tr.score,
// 	tr.score > th,
// 	tr.score > td
// 	{
//         padding: 10px;
//         border: 1px solid black;
//         border-collapse: collapse;
//       }

// 	table.sum,
// 	tr.sum,
// 	tr.sum > th,
// 	tr.sum > td
// 	{
// 		border: none;
// 		border-bottom: 0;
//         padding: 10px;
//       }

// 	</style>
// 	<script src=""></script>
// 	<body>`, season, round))

// 	for i, match := range matches {
// 		if i != 0 {
// 			buffer.WriteString(`<p style="page-break-after: always;">&#160;</p>`)
// 		}
// 		hteam := match.Home
// 		hnr := vicyear.ClubNummer(nil, hteam.Name)

// 		rteam := match.Remote
// 		rnr := vicyear.ClubNummer(nil, rteam.Name)
// 		round = match.Round
// 		actives := vicyear.Actives(nil, round)
// 		buffer.WriteString(fmt.Sprintf(`<table>
// <tr><td>Seizoen</td><td>%s</td></tr>
// <tr><td>Ronde</td><td>%s</td></tr>
// <tr><td>Afdeling</td><td>%s</td></tr>
// <tr><td>Datum</td><td>%s</td></tr>
// <tr><td>Teams</td><td>%s</td></tr>
// </table>
// `, match.Season, strings.ReplaceAll(round, "R", ""), hteam.Division, match.Date.Format(time.RFC3339)[:10], hteam.Name+" ("+hnr+")"+" vs. "+rteam.Name+" ("+rnr+")"))
// 		maxn := 0
// 		maxs := 0
// 		data := make([][8]string, 0)
// 		sumscore := "0-0"
// 		reverse := false
// 		if strings.HasPrefix(hteam.Name, "LANDEGEM") {
// 			duels := actives[hteam.Name]
// 			for i = 0; i < len(duels); i++ {
// 				j := strconv.Itoa(i + 1)
// 				duel := duels[j]
// 				ph := duel.Home
// 				pr := duel.Remote
// 				sumscore = sum(sumscore, duel.Score)
// 				if pr.Elo == "0" || pr.Elo == "" {
// 					pr.Elo = "-"
// 				}
// 				if ph.Elo == "0" || ph.Elo == "" {
// 					ph.Elo = "-"
// 				}
// 				data = append(data, [8]string{j, ph.Elo, ph.Name, duel.Score, pr.Name, pr.Elo, ph.Stam, pr.Stam})
// 				if len(ph.Elo) > maxs {
// 					maxs = len(ph.Elo)
// 				}
// 				if len(pr.Elo) > maxs {
// 					maxs = len(pr.Elo)
// 				}
// 				if len(ph.Name) > maxn {
// 					maxn = len(ph.Name)
// 				}
// 				if len(pr.Name) > maxn {
// 					maxn = len(pr.Name)
// 				}
// 			}
// 		} else {
// 			reverse = true
// 			duels := actives[rteam.Name]
// 			for i = 0; i < len(duels); i++ {
// 				j := strconv.Itoa(i + 1)
// 				duel := duels[j]
// 				ph := duel.Home
// 				pr := duel.Remote
// 				sumscore = sum(sumscore, duel.Score)
// 				if pr.Elo == "0" || pr.Elo == "" {
// 					pr.Elo = "-"
// 				}
// 				if ph.Elo == "0" || ph.Elo == "" {
// 					ph.Elo = "-"
// 				}
// 				if len(ph.Elo) > maxs {
// 					maxs = len(ph.Elo)
// 				}
// 				if len(pr.Elo) > maxs {
// 					maxs = len(pr.Elo)
// 				}
// 				if len(ph.Name) > maxn {
// 					maxn = len(ph.Name)
// 				}
// 				if len(pr.Name) > maxn {
// 					maxn = len(pr.Name)
// 				}
// 				data = append(data, [8]string{j, pr.Elo, pr.Name, duel.Score, ph.Name, ph.Elo, pr.Stam, ph.Stam})
// 			}
// 		}
// 		escape := html.EscapeString
// 		buffer.WriteString(`<p>&#160;</p><table class="score">`)
// 		buffer.WriteString(fmt.Sprintf(`<tr class="score"><th align="right" >Bord</th><th align="center">ELO <br />%s</th><th align="right">Speler <br />%s</th><th align="center">Score</th><th align="left">Speler <br />%s</th><th align="left">ELO <br />%s</th></tr>`, escape(hteam.Name), escape(hteam.Name), escape(rteam.Name), escape(rteam.Name)))

// 		for _, line := range data {
// 			escape := func(nr int) string { return html.EscapeString(line[nr]) }
// 			buffer.WriteString(fmt.Sprintf(`<tr class="score"><td align="right">%s</td><td class="clipboard" data-clip="%s" align="right" style="min-width:%dex;">%s</td><td class="clipboard"  align="right" style="min-width:%dex;">%s</td><td class="clipboard" align="center" style="min-width:3em;">%s</td><td class="clipboard" align="left" style="min-width:%dex;">%s</td><td data-clip="%s" class="clipboard" align="left" style="min-width:%dex;">%s</td></tr>`, escape(0), escape(6), maxs, escape(1), maxn, escape(2), escape(3), maxn, escape(4), escape(7), maxs, escape(5)))
// 		}
// 		buffer.WriteString(fmt.Sprintf(`<tr class="sum"><td align="right"></td><td align="right" ></td><td align="right"></td><td class="clipboard" align="center" style="min-width:5em;"><b>%s</b></td><td></td><td></td></tr>`, color(sumscore, reverse)))
// 		buffer.WriteString(`</table>`)
// 	}
// 	buffer.WriteString(`<script>
// function toClipboard() {
// 	var copyText = this.getAttribute('data-clip');
// 	if (!copyText) {
// 	    copyText = this.innerHTML;
// 	}
// 	copyText = copyText.trim();
// 	navigator.clipboard.writeText(copyText).then(() => {
// 		/* clipboard successfully set */
// 		}, () => {
// 		/* clipboard write failed */
// 		});
// 	}
// var elements = document.getElementsByClassName("clipboard");
// for (let i in elements) {
// 	let elm = elements.item(i);
// 	elm.addEventListener('click', toClipboard);
// }
// </script>
// `)

// 	buffer.WriteString(`</body>
// </html>`)

// 	mode := "pdf"
// 	if Fhtml {
// 		mode = "html"
// 	}
// 	outputfile := vicyear.OutputFile(nil, round, "html")
// 	htmlfile := outputfile
// 	pdffile := ""

// 	bfs.Store(outputfile, buffer.Bytes(), "")

// 	if mode == "pdf" {
// 		target := strings.TrimSuffix(outputfile, ".html") + ".pdf"
// 		pdffile = target
// 		aconvertor := vregistry.Registry["convert"].(map[string]any)["html2pdf"].([]any)
// 		convertor := make([]string, 0)
// 		for _, piece := range aconvertor {
// 			convertor = append(convertor, strings.ReplaceAll(strings.ReplaceAll(piece.(string), "{source}", outputfile), "{target}", target))
// 		}
// 		ccmd := exec.Command(convertor[0], convertor[1:]...)
// 		err := ccmd.Run()
// 		if err != nil {
// 			panic(err)
// 		}
// 		outputfile = target
// 	}

// 	fmt.Println("\n\n" + outputfile)

// 	aviewer := vregistry.Registry["viewer"].(map[string]any)["pdf"].([]any)

// 	if Fhtml {
// 		aviewer = vregistry.Registry["viewer"].(map[string]any)["html"].([]any)
// 	}
// 	viewer := make([]string, 0)

// 	for _, piece := range aviewer {
// 		viewer = append(viewer, strings.ReplaceAll(piece.(string), "{file}", outputfile))
// 	}
// 	vcmd := exec.Command(viewer[0], viewer[1:]...)
// 	err = vcmd.Start()
// 	if err != nil {
// 		panic(err)
// 	}
// 	if Fmail && YesNo("Mail the score ?") {
// 		return mailscore(season, round, htmlfile, pdffile)
// 	}

//}

// func sum(sumscore string, score string) string {
// 	score = strings.ReplaceAll(score, " ", "")
// 	x1, x2, _ := strings.Cut(score, "-")
// 	s1, s2, _ := strings.Cut(sumscore, "-")
// 	x1 = strings.TrimSuffix(x1, " f")
// 	x2 = strings.TrimSuffix(x2, " f")
// 	s1 = strings.TrimSuffix(s1, " f")
// 	s2 = strings.TrimSuffix(s2, " f")
// 	half := "½"
// 	add1 := 0
// 	add2 := 0
// 	if strings.Contains(s1, half) {
// 		add1 += 1
// 		s1 = strings.ReplaceAll(s1, half, "")
// 	}
// 	if strings.Contains(s2, half) {
// 		add2 += 1
// 		s2 = strings.ReplaceAll(s2, half, "")
// 	}
// 	x1 = strings.TrimSuffix(x1, " f")
// 	x2 = strings.TrimSuffix(x2, " f")
// 	if x1 == x2 {
// 		add1 += 1
// 		add2 += 1
// 	} else {
// 		if x2 != "0" {
// 			add2 += 2
// 		}
// 		if x1 != "0" {
// 			add1 += 2
// 		}
// 	}
// 	is1, _ := strconv.Atoi(s1)
// 	is2, _ := strconv.Atoi(s2)
// 	is1 *= 2
// 	is2 *= 2
// 	is1 += add1
// 	is2 += add2

// 	if is1%2 == 0 {
// 		s1 = strconv.Itoa(is1 / 2)
// 	} else {
// 		s1 = strconv.Itoa((is1-1)/2) + half
// 	}
// 	if is2%2 == 0 {
// 		s2 = strconv.Itoa(is2 / 2)
// 	} else {
// 		s2 = strconv.Itoa((is2-1)/2) + half
// 	}
// 	if s1 == ("0" + half) {
// 		s1 = half
// 	}
// 	if s2 == ("0" + half) {
// 		s2 = half
// 	}

// 	return s1 + "-" + s2

// }

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

	err = bmail.Send(to, cc, bcc, subject, "", string(html), []string{pdffile}, "")
	return err
}
