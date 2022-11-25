package structure

import (
	"os"

	"github.com/gocarina/gocsv"
)

type Result struct {
	Season        string `csv:"season"`
	Round         string `csv:"round"`
	Division      string `csv:"division"`
	Date          string `csv:"date"`
	VSLHome       string `csv:"vsl_at_home"`
	TeamhName     string `csv:"home_team_name"`
	TeamhClubno   string `csv:"home_team_stamno"`
	TeamrName     string `csv:"remote_team_name"`
	TeamrClubno   string `csv:"remote_team_stamno"`
	Board         string `csv:"board"`
	PlayerhName   string `csv:"home_player_name"`
	PlayerhStamno string `csv:"home_player_stamno"`
	PlayerhColor  string `csv:"home_player_color"`
	PlayerhELO    string `csv:"home_player_elo"`
	PlayerrName   string `csv:"remote_player_name"`
	PlayerrStamno string `csv:"remote_player_stamno"`
	PlayerrColor  string `csv:"remote_player_color"`
	PlayerrELO    string `csv:"remote_player_elo"`
	ScoreCode     string `csv:"scorecode"`
	Score         string `csv:"score"`
	TotalScore    string `csv:"score_total"`
}

func CSVwrite(results []*Result, fname string) (err error) {
	csvFile, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer csvFile.Close()
	gocsv.TagSeparator = ";"
	err = gocsv.MarshalFile(&results, csvFile)
	return
}
