package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	btime "brocade.be/base/time"
)

var Fyear string
var Fday string
var Fmax int

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Calendar `gopblad`",
	Long:  "Calendar `gopblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `gopblad calendar myfile.pb`,
	RunE:    calendar,
}

func init() {
	calendarCmd.PersistentFlags().IntVar(&Fmax, "max", 52, "max in year")
	calendarCmd.PersistentFlags().StringVar(&Fyear, "year", "", "for year")
	calendarCmd.PersistentFlags().StringVar(&Fday, "day", "", "for first Monday")
	rootCmd.AddCommand(calendarCmd)
}

func calendar(cmd *cobra.Command, args []string) error {

	if Fmax != 52 && Fmax != 53 {
		return fmt.Errorf("wrong value for --max")
	}

	now := time.Now()
	if Fyear == "" {
		month := now.Month()
		if month > 2 {
			Fyear = strconv.Itoa(now.Year() + 1)
		} else {
			Fyear = strconv.Itoa(now.Year())
		}
	}
	if Fyear != strconv.Itoa(now.Year()+1) && Fyear != strconv.Itoa(now.Year()) {
		return fmt.Errorf("wrong value for --year")
	}
	rex := regexp.MustCompile("^[0-9]{4}-[0-9]{2}-[0-9]{2}$")
	if !rex.MatchString(Fday) {
		return fmt.Errorf("wrong value for --day: YYYY-MM-DD")
	}
	fday := btime.DetectDate(Fday)
	if fday == nil {
		return fmt.Errorf("wrong value for --day: YYYY-MM-DD")
	}
	if fday.Month() != time.Month(1) {
		return fmt.Errorf("day should be in januari")
	}
	if strconv.Itoa(fday.Year()) != Fyear {
		return fmt.Errorf("day should be in " + Fyear)
	}
	if fday.Weekday().String() != "Monday" {
		return fmt.Errorf("day should be a Monday")
	}

	days := make(map[string]map[string]string)
	days[Fyear] = make(map[string]string)
	day := *fday
	for i := 0; i < Fmax; i++ {
		j := fmt.Sprintf("%02d", i+1)
		days[Fyear][j] = btime.StringDate(&day, "I")
		day = day.AddDate(0, 0, 7)
	}

	data, _ := json.MarshalIndent(days, "", "    ")
	fmt.Println(string(data))
	fmt.Println("\nChange dates as necessary. \nDelete holiday dates.\n Add to " + os.Getenv("MY_REGISTRY"))
	return nil
}
