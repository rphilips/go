package icyear

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	bfs "brocade.be/base/fs"
	vregistry "brocade.be/vchess/lib/registry"
)

var ELO = make(map[string]Player)
var Clubs = make(map[string]string)

type CSV struct {
	Season     string
	Round      string
	Date       string
	Division   string
	HomeNr     string
	HomeName   string
	RemoteNr   string
	RemoteName string
	Board      string
	WhiteNr    string
	WhiteName  string
	WhiteElo   string
	WhiteScore string
	BlackNr    string
	BlackName  string
	BlackScore string
}

type Match struct {
	Round  string
	Date   time.Time
	Season string
	Home   Team
	Remote Team
}

type Duel struct {
	Home   Player
	Remote Player
	Score  string
}

type Team struct {
	Name     string
	Nr       string
	Division string
	PK       string
	Players  []Player
}

type Player struct {
	Name      string
	Stam      string
	BaseTeam  string
	BasePlace string
	Elo       string
}

var rxdivision = regexp.MustCompile(`^AFD[A-Z]*\s+([0-9][A-Z])\s+[^A-Z]([A-Z]+)`)
var rxround = regexp.MustCompile("^R[0-9]+$")
var rxpairing = regexp.MustCompile("[0-9][0-9]?-[0-9][0-9]?$")
var rxdate = regexp.MustCompile("^[0-9][0-9]?-[0-9][0-9]?-[0-9][0-9][0-9][0-9]$")
var rxnumbers = regexp.MustCompile("^[0-9]+$")

func PlayerComplete(date *time.Time, player Player) Player {
	if player.Stam == "" {
		return player
	}
	elo, n := Elo(date, player.Stam)
	if player.Elo == "" {
		player.Elo = elo
	}
	if player.Name == "" {
		player.Name = n
	}
	return player
}
func Season(date *time.Time) string {
	if date == nil {
		d := time.Now()
		date = &d
	}
	year := date.Year()
	if date.Month() < 7 {
		year -= 1
	}
	return fmt.Sprintf("%d-%d", year, year+1)
}

func FName(date *time.Time, key string) (fname string) {
	season := Season(date)
	return vregistry.Registry["season"].(map[string]any)[season].(map[string]any)[key].(string)
}

func Elo(date *time.Time, stamnr string) (string, string) {
	season := Season(date)
	id := season + "-" + stamnr
	p := ELO[id]
	if p.Elo != "" {
		return p.Elo, p.Name
	}
	// 25518;Philips, Richard;;BEL;1915;;1957/07/07;
	elofile := FName(date, "elo")
	f, err := os.Open(elofile)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for _, record := range records {
		if len(record) < 6 {
			continue
		}
		stamnr := strings.TrimSpace(record[0])
		if stamnr == "" {
			continue
		}
		_, err := strconv.Atoi(stamnr)
		if err != nil {
			continue
		}
		id := season + "-" + stamnr
		p := ELO[id]
		if p.Elo != "" {
			continue
		}
		full := strings.TrimSpace(record[1])
		name := full
		if strings.Contains(full, ",") {
			parts := make([]string, 0)
			pieces := strings.SplitN(full, ",", -1)
			for i := len(pieces) - 1; i >= 0; i-- {
				n := strings.TrimSpace(pieces[i])
				if n == "" {
					continue
				}
				parts = append(parts, n)
			}
			name = strings.Join(parts, " ")
		}
		if record[4] == "None" {
			record[4] = "0"
		}

		ELO[id] = Player{
			Name: name,
			Stam: stamnr,
			Elo:  strings.TrimSpace(record[4]),
		}
	}
	id = season + "-" + stamnr
	return ELO[id].Elo, ELO[id].Name
}

func ClubNummer(date *time.Time, prefix string) string {
	prefix = strings.TrimRight(strings.ToUpper(strings.TrimSpace(prefix)), "1234567890 -")
	if prefix == "" {
		return ""
	}
	id := prefix
	nr := Clubs[id]
	if nr != "" {
		return nr
	}
	fname := FName(date, "clubs")
	data, err := os.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	lines := strings.SplitN(string(data), "\n", -1)
	for _, line := range lines {
		line := strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", -1)
		nr := strings.TrimSpace(parts[0])
		if nr == "" {
			continue
		}
		_, err := strconv.Atoi(nr)
		if err != nil {
			continue
		}
		name := strings.ToUpper(strings.TrimSpace(parts[1]))
		if strings.HasPrefix(name, prefix) {
			Clubs[id] = nr
		}
	}
	return Clubs[id]
}

func ActiveRound(date *time.Time, round string) string {
	s := FName(date, "active-round")
	return strings.ReplaceAll(s, "{round}", round)
}

func OutputFile(date *time.Time, round string, mode string) string {
	s := FName(date, mode+"-round")
	return strings.ReplaceAll(s, "{round}", round)
}

func AllTeams(records [][]string, date *time.Time) (teams []Team) {
	if len(records) == 0 {
		records, _ = Load(date)
	}
	division := ""
	pk := ""
	last := 0
	for _, record := range records {
		if len(record) < 2 {
			continue
		}
		field2 := strings.TrimSpace(strings.ToUpper(record[1]))
		if field2 == "" {
			continue
		}
		subs := rxdivision.FindAllStringSubmatch(field2, -1)
		if len(subs) != 0 {
			division = subs[0][1]
			pk = subs[0][2]
			last = 0
			continue
		}
		field1 := strings.TrimSpace(record[0])
		nr, ok := strconv.Atoi(field1)
		if ok != nil {
			continue
		}
		if nr != last+1 {
			continue
		}
		last = nr
		team := Team{
			Name:     field2,
			Nr:       field1,
			Division: division,
			PK:       pk,
		}
		teams = append(teams, team)
	}
	return
}

func Teams(records [][]string, date *time.Time) (teams []Team) {
	allteams := AllTeams(records, date)
	teams = make([]Team, 0)
	prefixen := vregistry.Registry["prefixen"].([]any)
	for _, team := range allteams {
		name := team.Name
		for _, n := range prefixen {
			n := strings.ToUpper(n.(string))
			if strings.HasPrefix(name, n) {
				teams = append(teams, team)
				break
			}
		}
	}

	return
}

func Matches(records [][]string, date *time.Time, iround string) []Match {
	if len(records) == 0 {
		records, _ = Load(date)
	}
	season := Season(date)

	allteams := AllTeams(records, date)
	teams := Teams(records, date)

	if date == nil {
		d := time.Now()
		date = &d
	}

	matches := make([]Match, 0)
	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		field1 := strings.TrimSpace(record[0])
		field1 = strings.ToUpper(field1)
		field1 = strings.ReplaceAll(field1, " ", "")
		if field1 == "" {
			continue
		}
		round := field1
		if !rxround.MatchString(field1) {
			continue
		}
		field2 := strings.TrimSpace(record[1])
		field2 = strings.ReplaceAll(field2, " ", "")
		field2 = strings.ReplaceAll(field2, ".", "-")
		field2 = strings.ReplaceAll(field2, "/", "-")
		if field2 == "" {
			continue
		}
		if !rxdate.MatchString(field2) {
			continue
		}
		parts := strings.SplitN(field2, "-", -1)
		year, _ := strconv.Atoi(parts[2])
		month, _ := strconv.Atoi(parts[1])
		day, _ := strconv.Atoi(parts[0])
		datum := time.Date(year, time.Month(month), day, 0, 0, 0, 0, date.Location())
		sdate := date.Format(time.RFC3339)[:10]
		sdatum := datum.Format(time.RFC3339)[:10]
		ok := round == iround || sdatum >= sdate
		if !ok {
			continue
		}
		for _, field := range record[2:] {
			field = strings.TrimSpace(field)
			field = strings.ReplaceAll(field, " ", "")
			field = strings.ReplaceAll(field, ".", "-")
			field = strings.ReplaceAll(field, "/", "-")
			if !rxpairing.MatchString(field) {
				continue
			}
			parts := strings.SplitN(field, "-", -1)
			home := parts[0]
			remote := parts[1]
			for _, t := range teams {
				if home == t.Nr || remote == t.Nr {
					bmatch := new(Match)
					bmatch.Date = datum
					bmatch.Round = round
					bmatch.Season = season
					if home == t.Nr {
						bmatch.Home = t
						for _, at := range allteams {
							if at.Nr == remote && t.Division == at.Division {
								bmatch.Remote = at
								break
							}
						}
					} else {
						bmatch.Remote = t
						for _, at := range allteams {
							if at.Nr == home && t.Division == at.Division {
								bmatch.Home = at
								break
							}
						}
					}
					matches = append(matches, *bmatch)
				}
			}
		}
		break
	}

	return matches
}

func Load(date *time.Time) (records [][]string, err error) {
	rooster := FName(date, "rooster")
	f, err := os.Open(rooster)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)
	r.Comma = ';'
	records, err = r.ReadAll()
	if err != nil {
		panic(err)
	}
	return
}

func Players(date *time.Time) (players []Player) {
	playerslist := FName(date, "players")

	f, err := os.Open(playerslist)
	if err != nil {
		panic(err)
	}
	r := csv.NewReader(f)
	r.Comma = ';'
	records, err := r.ReadAll()
	if err != nil {
		panic(err)
	}
	for _, record := range records {
		if len(record) < 8 {
			continue
		}
		name := record[0]
		stam := record[1]
		base := record[7]
		name = strings.TrimSpace(name)
		stam = strings.TrimSpace(stam)
		base = strings.TrimSpace(base)
		baseplace := ""
		baseteam := ""
		if strings.Contains(base, ":") {
			parts := strings.SplitN(base, ":", -1)
			baseplace = strings.TrimSpace(parts[1])
			baseteam = strings.TrimSpace(parts[0])
		}

		if stam == "" || name == "" {
			continue
		}
		if !rxnumbers.MatchString(stam) {
			continue
		}
		players = append(players, Player{
			Name:      name,
			Stam:      stam,
			BaseTeam:  baseteam,
			BasePlace: baseplace,
		})
	}
	return
}

func Actives(date *time.Time, round string) map[string]map[string]Duel {
	actives := make(map[string]map[string]Duel)
	players := Players(date)

	for _, player := range players {
		baseplace := player.BasePlace
		baseteam := player.BaseTeam
		if baseteam == "" {
			continue
		}
		if baseplace == "" {
			continue
		}
		_, ok := actives[baseteam]
		if !ok {
			actives[baseteam] = make(map[string]Duel)
		}
		actives[baseteam][baseplace] = Duel{
			Home: player,
		}
	}
	if round != "" {
		activeround := ActiveRound(date, round)
		jdata, err := os.ReadFile(activeround)
		if err != nil {
			data := make(map[string][]map[string]string)
			for team, active := range actives {
				_, ok := data[team]
				if !ok {
					data[team] = make([]map[string]string, 0)
				}
				plys := data[team]
				for nr, duel := range active {
					player := duel.Home
					inr, _ := strconv.Atoi(nr)
					for len(plys) < inr {
						plys = append(plys, make(map[string]string))
					}
					plys[inr-1]["home"] = player.Name
					plys[inr-1]["remote"] = ""
					plys[inr-1]["score"] = "½-½"
				}
				data[team] = plys
			}
			b, _ := json.MarshalIndent(data, "", "    ")
			bfs.Store(activeround, b, "")
			return actives
		}
		data := make(map[string][]map[string]string)
		err = json.Unmarshal(jdata, &data)
		if err != nil {
			panic(err)
		}
		for t, plys := range data {
			for i, ply := range plys {
				ok := false
				name := ply["home"]
				for _, player := range players {
					if player.Name == name {
						place := strconv.Itoa(i + 1)
						actives[t][place] = Duel{
							Home: PlayerComplete(date, player),
							Remote: PlayerComplete(date, Player{
								Stam: ply["remote"],
							}),
							Score: ply["score"],
						}
						ok = true
						break
					}
				}
				if !ok {
					fmt.Fprintf(os.Stderr, "error: cannot find player %s\n", name)
					os.Exit(1)
				}
			}
		}
	}
	return actives
}
