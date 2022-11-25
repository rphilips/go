package structure

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	vregistry "brocade.be/vchess/lib/registry"
)

var AllTeams = make(map[string]*Team)

var rxdivision = regexp.MustCompile(`^AFD[A-Z]*\s+([0-9][A-Z])\s+[^A-Z]([A-Z]+)`)

// var rxround = regexp.MustCompile("^R[0-9]+$")
// var rxpairing = regexp.MustCompile("[0-9][0-9]?-[0-9][0-9]?$")
// var rxdate = regexp.MustCompile("^[0-9][0-9]?-[0-9][0-9]?-[0-9][0-9][0-9][0-9]$")
// var rxnumbers = regexp.MustCompile("^[0-9]+$")

type Team struct {
	Name        string
	Number      string
	Division    string
	Club        *Club
	Players     []*Player
	PK          *Player
	BasePlayers []*Player
	VSL         bool
}

func (team Team) String() string {
	return team.Name
}

func (team *Team) Find(season *Season, name string) (err error) {
	if len(AllTeams) == 0 {
		rooster := season.FName("rooster")
		f, err := os.Open(rooster)
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
		division := ""
		last := 0
		basename := vregistry.Registry["club"].(map[string]any)["basename"].(string)
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
			team := new(Team)
			team.Name = field2
			team.Number = field1
			team.Division = division
			team.VSL = strings.HasPrefix(team.Name, basename)

			AllTeams[team.Name] = team
			AllTeams[team.Number+" "+division] = team
		}
	}
	if name == "" {
		name = team.Name
	}
	if name == "" && team.Number != "" && team.Division != "" {
		name = team.Number + " " + team.Division
	}
	if name == "" {
		return
	}
	t := AllTeams[name]
	if t != nil {
		baseplayers, e := BasePlayers(season, t.Name)
		if e != nil {
			err = e
			return
		}
		p := new(Player)
		err = p.Find(season, "")
		if err != nil {
			return err
		}
		team.Name = t.Name
		team.Number = t.Number
		team.Division = t.Division
		team.Players = t.Players
		team.PK = t.PK
		club := new(Club)
		err = club.Find(season, t.Name)
		if err != nil {
			return
		}
		team.Club = club
		team.BasePlayers = baseplayers
		basename := vregistry.Registry["club"].(map[string]any)["basename"].(string)
		team.VSL = strings.HasPrefix(team.Name, basename)
		AllTeams[name] = team
		return
	}

	return fmt.Errorf("cannot find team with name `%s`", name)
}
