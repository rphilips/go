package structure

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	vregistry "brocade.be/vchess/lib/registry"
	vstrings "brocade.be/vchess/lib/strings"
)

type Season struct {
	Season string
}

func (season *Season) Init(date *time.Time) {
	if date == nil {
		x := time.Now()
		date = &x
	}
	year := date.Year()
	if date.Month() < 7 {
		year -= 1
	}
	season.Season = fmt.Sprintf("%d-%d", year, year+1)
}

func (season Season) String() string {
	return season.Season
}

func (season Season) FName(key string) (fname string) {
	s := vregistry.Registry["season"].(map[string]any)[season.Season].(map[string]any)[key].(string)
	return strings.ReplaceAll(s, "{season}", season.String())
}

func (season Season) HomeTeams() (teams []*Team, err error) {
	team := new(Team)
	err = team.Find(&season, "")
	for _, team := range AllTeams {
		if team.VSL {
			teams = append(teams)
		}
	}
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].Name < teams[j].Name
	})

	return
}

func (season Season) Teams() (teams []*Team, err error) {
	team := new(Team)
	err = team.Find(&season, "")
	for _, team := range AllTeams {
		teams = append(teams, team)
	}
	return
}

func (season *Season) Round(roundnr int) (round *Round, err error) {
	rooster := season.FName("rooster")
	f, err := os.Open(rooster)
	if err != nil {
		return
	}
	csvr := csv.NewReader(f)
	csvr.Comma = ';'
	csvr.FieldsPerRecord = -1
	rrecords, err := csvr.ReadAll()
	if err != nil {
		return
	}

	if roundnr == 0 {
		date := time.Now()
		for _, rr := range rrecords {
			if len(rrecords) < 2 {
				continue
			}
			field1 := strings.ToUpper(strings.TrimSpace(rr[0]))
			if !strings.HasPrefix(field1, "R") {
				continue
			}
			field1 = strings.TrimSpace(strings.TrimLeft(field1, "R"))
			if field1 == "" {
				continue
			}
			i, e := strconv.Atoi(field1)
			if e != nil {
				continue
			}
			d := strings.TrimSpace(rr[1])
			t := vstrings.DetectDate(d)
			if t == nil {
				continue
			}
			if vstrings.StringDate(t, "I") < vstrings.StringDate(&date, "I") {
				continue
			}
			roundnr = i
			break
		}
	}
	if roundnr == 0 {
		err = fmt.Errorf("cannot find round")
		return
	}

	who := make([]string, 0)
	for _, rr := range rrecords {
		if len(rr) < 2 {
			continue
		}
		field1 := strings.ToUpper(strings.TrimSpace(rr[0]))
		if !strings.HasPrefix(field1, "R") {
			continue
		}
		field1 = strings.TrimSpace(strings.TrimLeft(field1, "R"))
		if field1 == "" {
			continue
		}
		i, e := strconv.Atoi(field1)
		if e != nil {
			continue
		}
		if i != roundnr {
			continue
		}
		who = rr[1:]
	}
	if len(who) < 2 {
		round = nil
		err = fmt.Errorf("cannot find round %d", roundnr)
		return
	}
	d := strings.TrimSpace(who[0])
	t := vstrings.DetectDate(d)
	if t == nil {
		err = fmt.Errorf("cannot find date with round %d", roundnr)
		return
	}

	round = new(Round)
	round.Round = roundnr
	round.Date = t
	division := ""
	for _, rr := range rrecords {
		if len(rr) < 2 {
			continue
		}
		for _, f := range rr {
			f := strings.TrimSpace(f)
			if f == "" {
				continue
			}
			if strings.HasPrefix(f, "AFDELING") {
				f = strings.TrimSpace(strings.TrimPrefix(f, "AFDELING"))
				k := strings.IndexAny(f, "( ")
				if k != -1 {
					f = strings.TrimSpace(f[:k])
				}
				if f == "" {
					continue
				}
				division = f
				for _, match := range who[1:] {
					match := strings.TrimSpace(match)
					if !strings.Contains(match, "-") {
						continue
					}
					x, y, _ := strings.Cut(match, "-")
					x = strings.TrimSpace(x)
					y = strings.TrimSpace(y)
					if x == "" || y == "" {
						continue
					}
					_, e := strconv.Atoi(x)
					if e != nil {
						continue
					}
					_, e = strconv.Atoi(y)
					if e != nil {
						continue
					}
					teamh := new(Team)
					teamh.Division = division
					teamh.Number = x

					teamh.Find(season, "")
					teamr := new(Team)
					teamr.Division = division
					teamr.Number = y
					teamr.Find(season, "")
					if !teamh.VSL && !teamr.VSL {
						continue
					}
					if teamh.Name == "" || teamr.Name == "" {
						continue
					}
					duel := new(Duel)
					duel.Home = teamh
					duel.Remote = teamr
					duel.Score = "0-0"
					round.Duels = append(round.Duels, duel)
				}
			}
		}
	}

	jround := strings.ReplaceAll(season.FName("active-round"), "{round}", strconv.Itoa(roundnr))
	jdata, enf := os.ReadFile(jround)
	if enf != nil {
		m := make(map[string][]map[string]string)
		for _, duel := range round.Duels {
			var team *Team = nil
			if duel.Home.VSL {
				team = duel.Home
				duel.Home.Players = append([]*Player{}, duel.Home.BasePlayers...)
			} else {
				team = duel.Remote
				duel.Remote.Players = append([]*Player{}, duel.Remote.BasePlayers...)
			}
			score := "½-½"
			m[team.Name] = make([]map[string]string, 0)
			for _, p := range team.Players {
				m[team.Name] = append(m[team.Name], map[string]string{
					"vsl":   p.Name,
					"other": "",
					"score": score,
				})
			}
		}
		b, _ := json.MarshalIndent(m, "", "    ")
		bfs.Store(jround, b, "")
		jdata, err = os.ReadFile(jround)
		if err != nil {
			return
		}
	}
	m := make(map[string][]map[string]string)
	err = json.Unmarshal(jdata, &m)
	if err != nil {
		return
	}
	for _, duel := range round.Duels {
		teamvsl := new(Team)
		teamoth := new(Team)
		seq := true
		if duel.Home.VSL {
			teamvsl = duel.Home
			teamoth = duel.Remote
		} else {
			teamvsl = duel.Remote
			teamoth = duel.Home
			seq = false
		}
		name := teamvsl.Name
		if name == "" {
			err = fmt.Errorf("team with empty name")
			return
		}
		plays, ok := m[name]
		if !ok {
			err = fmt.Errorf("team with name `%s` not found", name)
			return
		}
		score := "0-0"
		for _, play := range plays {
			vslplayer := play["vsl"]
			if vslplayer == "" {
				err = fmt.Errorf("player with empty name in team `%s`", name)
				return
			}
			p := new(Player)
			p.Find(season, vslplayer)
			if p.Stamno == "" {
				err = fmt.Errorf("player `%s` has no stamno", vslplayer)
				return
			}
			teamvsl.Players = append(teamvsl.Players, p)
			stamno := play["other"]
			q := new(Player)
			if stamno != "" {
				e := q.Find(season, stamno)
				if e != nil {
					err = fmt.Errorf("player with stamno `%s` is unknown", stamno)
					return
				}
			}
			teamoth.Players = append(teamoth.Players, q)
			match := new(Match)
			match.VSL = p
			match.Other = q
			match.Score = play["score"]
			duel.Match = append(duel.Match, match)
			score = sum(score, match.Score)
		}
		seq = true
		if !seq {
			x1, x2, _ := strings.Cut(score, "-")
			score = x2 + "-" + x1
		}
		duel.Score = score
	}
	return
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

func (season *Season) OutputFile(round string, mode string, name string) string {
	mode = mode + "-round"
	if name != "" {
		mode += "-desktop"
	}
	round = strings.ReplaceAll(round, "R", "")
	s := season.FName(mode)
	return strings.ReplaceAll(s, "{round}", round)
}

func (season *Season) CalendarFile(mode string) string {
	s := season.FName(mode + "-calendar")
	return s
}

func (season *Season) ClubFile(clubno string) string {
	s := season.FName("club-round")
	s = strings.ReplaceAll(s, "{clubno}", clubno)
	return s
}

func (season *Season) Results() (fname string, lastround int, results []*Result, err error) {
	fname = season.FName("csv-results")
	fround := season.FName("active-round")
	for i := 1; ; i++ {
		f := strings.ReplaceAll(fround, "{round}", strconv.Itoa(i))
		data, e := os.ReadFile(f)
		if e != nil {
			break
		}
		m := make(map[string]any)
		err = json.Unmarshal(data, &m)
		if err != nil {
			return
		}

		round, e := season.Round(i)
		if e != nil {
			err = e
			return
		}

		for _, duel := range round.Duels {
			teamh := duel.Home
			teamr := duel.Remote
			vslhome := "0"
			if duel.Home.VSL {
				vslhome = "1"
			}

			colorh := "black"
			colorr := "white"
			for j, match := range duel.Match {
				if match.Other.Stamno == "" {
					continue
				}
				lastround = i
				p := match.VSL
				q := match.Other
				if duel.Remote.VSL {
					p = match.Other
					q = match.VSL
				}
				if colorh == "black" {
					colorh = "white"
					colorr = "black"
				} else {
					colorr = "white"
					colorh = "black"
				}
				result := Result{
					Season:      season.String(),
					Round:       round.String(),
					Division:    teamh.Division,
					Date:        round.Date.Format(time.RFC3339)[:10],
					VSLHome:     vslhome,
					TeamhName:   teamh.String(),
					TeamhClubno: teamh.Club.Stamno,
					TeamrName:   teamr.String(),
					TeamrClubno: teamr.Club.Stamno,

					Board:         strconv.Itoa(j + 1),
					PlayerhName:   p.Name,
					PlayerhStamno: p.Stamno,
					PlayerhColor:  colorh,
					PlayerhELO:    p.Elo,
					PlayerrName:   q.Name,
					PlayerrStamno: q.Stamno,
					PlayerrColor:  colorr,
					PlayerrELO:    q.Elo,
					ScoreCode:     "",
					Score:         match.Score,
					TotalScore:    duel.Score,
				}
				results = append(results, &result)
			}
		}

	}
	err = CSVwrite(results, fname)
	return
}

func (season *Season) Calendar() (rounds []*Round, err error) {
	// if date == nil {
	// 	d := time.Now()
	// 	date = &d
	// }

	r := 0
	for {
		r++
		round, err := season.Round(r)
		if round == nil || err != nil {
			break
		}
		rounds = append(rounds, round)
	}

	escape := html.EscapeString
	buf := make([]byte, 0)
	buffer := bytes.NewBuffer(buf)
	buffer.WriteString(fmt.Sprintf(`<!DOCTYPE html>
			<html lang="nl">
			<meta charset="UTF-8">
			<title>%s</title>
			<style>
			table,
			tr.before,
			tr.before > th,
			tr.before > td
			{
		        padding: 10px;
		        border: 1px solid black;
		        border-collapse: collapse;
				font-style: italic;
				font-size: smaller;
		      }
			  tr.after,
			tr.after > th,
			tr.after > td
			{
		        padding: 10px;
		        border: 1px solid black;
		        border-collapse: collapse;
				font-style: normal;
		      }

			</style>
			<script src=""></script>
			<body>

			<table>`, season))

	now := time.Now()
	for _, round := range rounds {
		class := "before"
		if !round.Date.Before(now) {
			class = "after"
		}
		buffer.WriteString(fmt.Sprintf(`<tr class="%s"><td class="%s">%s</td><td class="%s">%s</td><td class="%s">`, class, class, round, class, vstrings.StringDate(round.Date, ""), class))
		for _, duel := range round.Duels {
			teamh := duel.Home
			teamr := duel.Remote
			nameh := "<b>" + escape(teamh.String()) + "</b>"
			if !teamh.VSL {
				nameh = escape(teamh.String())
			}
			namer := "<b>" + escape(teamr.String()) + "</b>"
			if !teamr.VSL {
				namer = escape(teamr.String())
			}
			if teamh.Club == nil {
				err = fmt.Errorf("cannot find a club for `%s`", teamh)
				return
			}
			if teamr.Club == nil {
				err = fmt.Errorf("cannot find a club for `%s`", teamr)
				return
			}
			buffer.WriteString(nameh + " (" + teamh.Club.Stamno + ") vs. " + namer + " (" + teamr.Club.Stamno + ")<br />")
		}
		buffer.WriteString(`</td></tr>`)
	}
	buffer.WriteString(`</table></body></html>`)
	calfile := season.CalendarFile("html")
	bfs.Store(calfile, buffer.Bytes(), "")
	target := strings.TrimSuffix(calfile, ".html") + ".pdf"
	aconvertor := vregistry.Registry["convert"].(map[string]any)["html2pdf"].([]any)
	convertor := make([]string, 0)
	for _, piece := range aconvertor {
		convertor = append(convertor, strings.ReplaceAll(strings.ReplaceAll(piece.(string), "{source}", calfile), "{target}", target))
	}
	ccmd := exec.Command(convertor[0], convertor[1:]...)
	err = ccmd.Run()
	return
}
