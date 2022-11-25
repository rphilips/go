package structure

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Player struct {
	Name      string
	Stamno    string
	BaseTeam  string
	BasePlace int
	Email     string
	Teams     []string
	Elo       string
	PK        *Team
	VSL       bool
}

var AllPlayers = make(map[string]*Player)

func (player *Player) Find(season *Season, stamno string) (err error) {
	if len(AllPlayers) == 0 {
		// all players
		elofile := season.FName("elo")
		f, err := os.Open(elofile)
		if err != nil {
			return err
		}
		r := csv.NewReader(f)
		r.Comma = ';'
		r.FieldsPerRecord = -1
		records, err := r.ReadAll()
		if err != nil {
			return err
		}
		f.Close()
		for _, record := range records {
			if len(record) < 6 {
				continue
			}
			stamno := strings.TrimSpace(record[0])
			if stamno == "" {
				continue
			}
			_, err := strconv.Atoi(stamno)
			if err != nil {
				continue
			}
			player := new(Player)
			player.Stamno = stamno
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
			player.Name = name
			elo := strings.TrimSpace(record[4])
			_, e := strconv.Atoi(elo)
			if e != nil {
				elo = "0"
			}

			player.Elo = elo
			AllPlayers[stamno] = player
		}
		// vsl players
		vslfile := season.FName("players")
		f, err = os.Open(vslfile)
		if err != nil {
			return err
		}
		r = csv.NewReader(f)
		r.Comma = ';'
		r.FieldsPerRecord = -1
		records, err = r.ReadAll()
		if err != nil {
			return err
		}
		f.Close()
		for _, record := range records {
			if len(record) < 8 {
				continue
			}
			stamno := strings.TrimSpace(record[1])
			_, e := strconv.Atoi(stamno)
			if e != nil {
				continue
			}
			player := AllPlayers[stamno]
			if player == nil {
				return fmt.Errorf("homeplayer with stamno `%s` is not known in KBSB", stamno)
			}
			player.VSL = true
			base := strings.TrimSpace(record[7])
			place := ""
			if strings.Contains(base, ":") {
				parts := strings.SplitN(base, ":", -1)
				place = strings.TrimSpace(parts[1])
				base = strings.TrimSpace(parts[0])
			}
			if base != "" {
				bplace, e := strconv.Atoi(place)
				if e != nil || bplace < 1 {
					return fmt.Errorf("homeplayer with stamno `%s` has wrong order in team", stamno)
				}
				player.BaseTeam = base
				player.BasePlace = bplace
			}

			name := strings.TrimSpace(record[0])
			if name != "" {
				AllPlayers[name] = player
			}

			player.Email = strings.TrimSpace(record[10])
			teampk := strings.TrimSpace(record[9])
			if teampk != "" {
				team := new(Team)
				team.Find(season, teampk)
				if team.Name != teampk {
					return fmt.Errorf("homeplayer with stamno `%s` is captain of wrong team", stamno)
				}
				player.PK = team
			}
			teams := strings.TrimSpace(record[8])
			if teams != "" {
				parts := strings.SplitN(teams, "+", -1)
				for _, part := range parts {
					part := strings.TrimSpace(part)
					if part == "" {
						continue
					}
					player.Teams = append(player.Teams, part)
				}
			}
		}
	}
	if stamno == "" {
		stamno = player.Stamno
	}
	if stamno == "" {
		return
	}
	x := AllPlayers[stamno]
	if x != nil {
		player.Name = x.Name
		player.Stamno = x.Stamno
		player.BaseTeam = x.BaseTeam
		player.BasePlace = x.BasePlace
		player.Email = x.Email
		player.Teams = append([]string{}, x.Teams...)
		player.Elo = x.Elo
		player.VSL = x.VSL
		return
	}
	return fmt.Errorf("player with stamno `%s` is not known", stamno)
}

func BasePlayers(season *Season, teamname string) (baseteam []*Player, err error) {
	p := new(Player)
	err = p.Find(season, "")
	if err != nil {
		return
	}
	found := make(map[string]bool)
	for _, p := range AllPlayers {
		if p.BaseTeam != teamname {
			continue
		}

		if found[p.Stamno] {
			continue
		}
		found[p.Stamno] = true
		baseteam = append(baseteam, p)
	}
	sort.Slice(baseteam, func(i, j int) bool {
		return baseteam[i].BasePlace < baseteam[j].BasePlace
	})
	return
}
